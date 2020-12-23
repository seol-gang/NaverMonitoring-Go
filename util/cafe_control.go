package util

import (
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NewControlData function
// 카페 컨트롤 생성자
func NewControlData(port int, channel chan []ArticleID) *ControlData {
	return &ControlData{D: NewDriver(port), articleUrl: channel}
}

// contains function
// 특정 게시판 확인 함수
func contains(s []string, substr string) bool {
	for _, v := range s {
		if v == substr {
			return true
		}
	}
	return false
}

func (d ControlData) FilterArticle() {
	var wait sync.WaitGroup
	wait.Add(1)
	go func(){
		defer wait.Done()
		d.D.LoopUpdateMessage()
	}()


	var articleInfo []ArticleID
	for {
		articleInfo = <- d.articleUrl
		for _, data := range articleInfo {
			d.D.Driver.Get(data.link)
			time.Sleep(2 * time.Second)
			if err := d.D.Driver.AcceptAlert(); err == nil {
				continue
			}
			if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, "#cafe_main")); err != nil {
				fmt.Print(err)
			}
			if err := d.D.Driver.SwitchFrame("cafe_main"); err != nil {
				fmt.Println(err)
			}
			if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, ".article_viewer")); err != nil {
				fmt.Println(err)
			}

			if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, ".link_board")); err != nil {
				fmt.Println(err)
			}
			articleTypeElement, _ := d.D.Driver.FindElement(selenium.ByCSSSelector, ".link_board")
			articleType, _ := articleTypeElement.Text() // 게시판 종류

			if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, ".title_text")); err != nil {
				fmt.Println(err)
			}
			articleTitleElement, _ := d.D.Driver.FindElement(selenium.ByCSSSelector, ".title_text")
			articleTitle, _ := articleTitleElement.Text() // 게시글 제목

			articleTextElement, err := d.D.Driver.FindElement(selenium.ByCSSSelector, "div.content.CafeViewer")
			if err != nil {
				articleTextElement, _ = d.D.Driver.FindElement(selenium.ByCSSSelector, ".ContentRenderer")
			}
			var articleText string
			for {
				time.Sleep(1 * time.Second)
				articleText, err = articleTextElement.GetAttribute("innerHTML") // 게시글 내용
				if err == nil {
					break
				}
			}

			if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, ".ArticleWriterProfile")); err != nil {
				fmt.Println(err)
			}
			writerIdElement, _ := d.D.Driver.FindElement(selenium.ByCSSSelector, ".ArticleWriterProfile a")
			writerID, _ := writerIdElement.GetAttribute("href")
			writerID = strings.Split(writerID, "=")[len(strings.Split(writerID, "=")) - 1] // 게시글 작성자 ID

			writerNicknameElement, _ := d.D.Driver.FindElement(selenium.ByCSSSelector, ".user")
			writerNickname, _ := writerNicknameElement.Text()

			//게시글 제목 정규표현식 작성
			titleRegexp, _ := regexp.Compile("(대리|ㄷㄹ|머리|듀오|ㄷㅇ|기사|버스|강의|페이|염전|부주|빌리|빌려|빌림|승당|도움|도와|상승|패작|계정|판매|팜|팔아|ㅁㅁ|ㅍㅍ|카톡|디코|ㄷㅋ|ㄱㅎ|캐리|대1리|ㄷ1ㄹ|최저|바리|싸게)")
			//게시글 내용 정규표현식 작성
			contentRegexp, _ := regexp.Compile("(open.kakao.com|discord.gg)(\\?|\\/)?([a-zA-Z0-9-_&=])([^\\'\" >]+)|((카|톡|디)([1-9a-zA-Z\\s]+)?(톡|디|코))|(discord|디스코드|쪽지|일쳇|1:1|일대일|톡)")
			//게시글 고유 번호 정규표현식
			articleIdRegexp, _ := regexp.Compile("articleid=\\d+")

			titleMatchStr := titleRegexp.FindAllString(articleTitle, -1)
			contentMatchStr := contentRegexp.FindAllString(articleText, -1)
			articleIdStr := strings.Split(articleIdRegexp.FindString(data.link), "=")[1]

			if len(contentMatchStr) != 0 {
				var filterTitle string
				var filterContent string
				filterURL := "https://cafe.naver.com/lolkor/" + articleIdStr

				for _, data := range titleMatchStr {
					filterTitle += data
				}

				for _, data := range contentMatchStr {
					filterContent += data
				}

				message := fmt.Sprintf("[필터 감지]\n==========\n게시판 종류 : %s\n==========\n필터 제목 : %s\n==========\n필터 내용 : %s\n==========\n닉네임 : %s\n==========\n아이디 : %s\n==========\n링크 주소 : %s\n==========",
					articleType, filterTitle, filterContent, writerNickname, writerID, filterURL)

				d.D.SendBotMessage(message)
			}
		}
	}

	wait.Done()
}


// FindFilterArticle function
// 특정 게시판 특정 제목 필터링 함수
func (d ControlData) FindFilterArticle() {
	//필터할 특정 게시판 이름 목록 작성
	filterAritcleList := []string{
		"자유 게시판",
		"랭크 게임",
		"격전",
		"부캐 듀오",
		"일반/칼바람",
		"특수모드/경작/AI",
		"대회멤버/스크림",
		"기타 서버",
	}
	var currentArticleID int // 한 페이지 내 확인했던 글 다시 확인하는것을 방지
	currentArticleID = 0
	
	//필터링할 카페 전체게시글 게시판 URL 입력
	d.D.Driver.Get("https://cafe.naver.com/ArticleList.nhn?search.clubid=19543191&search.boardtype=L")
	for {
		var articleIdList []ArticleID // 마지막 작성글 ID를 가져오기 위함

		if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, "#cafe_main")); err != nil {
			fmt.Println(fmt.Sprintf("FindFilterArticle error : %s", err))
		}

		if err := d.D.Driver.SwitchFrame("cafe_main"); err != nil {
			fmt.Println(fmt.Sprintf("FindFilterArticle : %s", err))
		}

		t, err := d.D.Driver.FindElement(selenium.ByXPATH, "//*[@id=\"main-area\"]/div[4]/table/tbody")
		if err != nil {
			fmt.Println(err)
		}
		article, err := t.FindElements(selenium.ByXPATH, "//*[@id=\"main-area\"]/div[4]/table/tbody/tr")
		if err != nil {
			fmt.Println(err)
		}

		for _, data := range article {
			/////////////////
			// 게시판 목록 확인
			boardType, err := data.FindElement(selenium.ByClassName, "inner_name")
			if err != nil {
				fmt.Println(err)
			}
			articleBoardName, _ := boardType.Text()
			if !contains(filterAritcleList, articleBoardName) {
				continue
			}
			/////////////

			/////////////
			// 게시판 고유 ID 및 링크 확인
			articleIdObj := ArticleID{}

			articleElement, err := data.FindElement(selenium.ByCSSSelector, ".article")
			if err != nil {
				fmt.Println(err)
			}
			articleLink, err := articleElement.GetAttribute("href")
			if err != nil {
				fmt.Println(err)
			}
			r, _ := regexp.Compile("articleid=[0-9]+")
			temp := r.FindString(articleLink)
			temp = strings.Split(temp, "=")[1]
			articleID, _ := strconv.Atoi(temp)
			if articleID <= currentArticleID {
				break
			}
			articleIdObj.link = articleLink
			articleIdObj.id = articleID
			articleIdList = append(articleIdList, articleIdObj)
			////////////
		}
		fmt.Println(len(articleIdList))
		if len(articleIdList) != 0 {
			currentArticleID = articleIdList[0].id
			d.articleUrl <- articleIdList
		}
		time.Sleep(3 * time.Second)
		d.D.Driver.Refresh()
	}
}

// ChangeImageSrc function
// 게시글 내 이미지 src 링크를 바이너리 데이터로 바꿈
// 사용 보류
func (d ControlData)ChangeImageSrc(url string) string {
	d.D.Driver.Get(url)
	if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, "#cafe_main")); err != nil {
		fmt.Println(fmt.Sprintf("ChangeImageSrc error : %s", err))
	}
	if err := d.D.Driver.SwitchFrame("cafe_main"); err != nil {
		fmt.Println(err)
	}
	if err := d.D.Driver.Wait(Enabled(selenium.ByCSSSelector, ".article_container")); err != nil {
		fmt.Println(fmt.Sprintf("ChangeImageSrc error : %s", err))
	}

	html, _ := d.D.Driver.PageSource()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		fmt.Println(err)
	}
	doc.Find(".article_container").Each(func(i int, s *goquery.Selection) {
		imageUrl, _ := s.Find("img").Attr("src")
		fileExtension := strings.Split(filepath.Ext(imageUrl), "?")[0][1:]
		response, _ := http.Get(imageUrl)
		defer response.Body.Close()
		content, _ := ioutil.ReadAll(response.Body)
		encoded := base64.StdEncoding.EncodeToString(content)
		binarySrc := "data:image/" + fileExtension + ";base64," + encoded
		html = strings.Replace(html, imageUrl, binarySrc, 1)
	})
	return html
}
