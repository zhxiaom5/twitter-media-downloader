package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	URL "net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	twitterscraper "github.com/jeffrey12cali/twitter-scraper"
	"github.com/mmpx12/optionparser"
	"github.com/sirupsen/logrus"
)

var (
	usr     string
	format  string
	proxy   string
	update  bool
	onlyrtw bool
	onlymtw bool
	vidz    bool
	imgs    bool
	urlOnly bool
	version = "1.15.0"
	scraper *twitterscraper.Scraper
	client  *http.Client
	size    = "orig"
	datefmt = "2006-01-02"

	// Logger instance
	logger = logrus.New()
)

func init() {
	// Configure logger
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

func generateNFOFile(tweet interface{}, videoUrl string, output string, dwn_type string) {
	// Generate nfo filename (same as video but with .nfo extension)
	segments := strings.Split(videoUrl, "/")
	videoName := segments[len(segments)-1]
	re := regexp.MustCompile(`name=`)
	if re.MatchString(videoName) {
		segments := strings.Split(videoName, "?")
		videoName = segments[len(segments)-2]
	}

	// Get tweet content for filename
	tweetContent := "没有推文"
	pattern := `[/\\:*?"<>|]`
	regex, _ := regexp.Compile(pattern)

	// Extract tweet information
	var title, description, author, date string
	var tweetID string

	// Log tweet processing
	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		tweetID = t.ID
	case *twitterscraper.Tweet:
		tweetID = t.ID
	}
	logger.Infof("Processing tweet: %s", tweetID)

	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
			description = t.Text
		} else {
			description = "没有推文"
		}
		title = t.Name + "的推文"
		author = t.Username
		date = time.Unix(t.Timestamp, 0).Format("2006-01-02")
		tweetID = t.ID
	case *twitterscraper.Tweet:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
			description = t.Text
		} else {
			description = "没有推文"
		}
		title = t.Name + "的推文"
		author = t.Username
		date = time.Unix(t.Timestamp, 0).Format("2006-01-02")
		tweetID = t.ID
	}

	// Create nfo filename
	nameWithoutExt := strings.TrimSuffix(videoName, "."+strings.Split(videoName, ".")[len(strings.Split(videoName, "."))-1])
	nfoName := nameWithoutExt + "_" + tweetContent + ".nfo"

	// Create nfo file path
	var nfoPath string
	if dwn_type == "user" {
		if _, err := os.Stat(output + "/video"); os.IsNotExist(err) {
			os.MkdirAll(output+"/video", os.ModePerm)
		}
		nfoPath = output + "/video/" + nfoName
	} else {
		if _, err := os.Stat(output); os.IsNotExist(err) {
			os.MkdirAll(output, os.ModePerm)
		}
		nfoPath = output + "/" + nfoName
	}

	// Generate nfo content (Jellyfin compatible XML)
	nfoContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<movie>
  <title>%s</title>
  <originaltitle>%s</originaltitle>
  <plot>%s</plot>
  <outline>%s</outline>
  <year>%s</year>
  <premiered>%s</premiered>
  <aired>%s</aired>
  <studio>Twitter</studio>
  <director>%s</director>
  <credits>%s</credits>
  <actor>
    <name>%s</name>
    <role>作者</role>
  </actor>
  <tag>Twitter</tag>
  <tag>视频</tag>
  <uniqueid type="twitter" default="true">%s</uniqueid>
</movie>`,
		title, title, description, description, date[:4], date, date, author, author, author, tweetID)

	// Write nfo file
	err := os.WriteFile(nfoPath, []byte(nfoContent), 0644)
	if err == nil {
		logger.Infof("Generated NFO file: %s", nfoName)
		logger.Infof("NFO file path: %s", nfoPath)
	} else {
		logger.Errorf("Failed to generate NFO file: %s", err.Error())
	}
}

func generateASSFile(tweet interface{}, videoUrl string, output string, dwn_type string) {
	// Generate ass filename (same as video but with .ass extension)
	segments := strings.Split(videoUrl, "/")
	videoName := segments[len(segments)-1]
	re := regexp.MustCompile(`name=`)
	if re.MatchString(videoName) {
		segments := strings.Split(videoName, "?")
		videoName = segments[len(segments)-2]
	}

	// Get tweet content for subtitle
	tweetContent := "没有推文"
	pattern := `[/\\:*?\"<>|]`
	regex, _ := regexp.Compile(pattern)

	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
		}
	case *twitterscraper.Tweet:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
		}
	}

	// Create ass filename
	nameWithoutExt := strings.TrimSuffix(videoName, "."+strings.Split(videoName, ".")[len(strings.Split(videoName, "."))-1])
	assName := nameWithoutExt + "_" + tweetContent + ".ass"

	// Create ass file path
	var assPath string
	if dwn_type == "user" {
		if _, err := os.Stat(output + "/video"); os.IsNotExist(err) {
			os.MkdirAll(output+"/video", os.ModePerm)
		}
		assPath = output + "/video/" + assName
	} else {
		if _, err := os.Stat(output); os.IsNotExist(err) {
			os.MkdirAll(output, os.ModePerm)
		}
		assPath = output + "/" + assName
	}

	// Generate ass content
	assContent := fmt.Sprintf(`[Script Info]
; Script generated by twmd
Title: Twitter Video Subtitle
Original Script: twmd
ScriptType: v4.00+
Collisions: Normal
PlayResX: 1080
PlayResY: 1920
WrapStyle: 3
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H00000099,0,0,0,0,100,100,0,0,1,2,2,2,40,40,80,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:00.00,99:59:59.99,Default,,40,40,80,,%s`,
		tweetContent)

	// Write ass file
	err := os.WriteFile(assPath, []byte(assContent), 0644)
	if err == nil {
		logger.Infof("Generated ASS file: %s", assName)
		logger.Infof("ASS file path: %s", assPath)
	} else {
		logger.Errorf("Failed to generate ASS file: %s", err.Error())
	}
}

func downloadThumbnail(wg *sync.WaitGroup, tweet interface{}, video interface{}, videoUrl string, output string, dwn_type string) {
	defer wg.Done()

	// Log thumbnail download start
	logger.Infof("Starting thumbnail download")
	logger.Infof("Video URL: %s", videoUrl)
	logger.Infof("Output directory: %s", output)

	// Extract thumbnail URL from video object
	var thumbnailUrl string

	// First, try to get thumbnail URL from video object using reflection
	v := reflect.ValueOf(video)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try to get Preview field from video object
	previewField := v.FieldByName("Preview")
	if previewField.IsValid() {
		// Check if Preview is a struct with URL field
		if previewField.Kind() == reflect.Struct {
			urlField := previewField.FieldByName("URL")
			if urlField.IsValid() && urlField.Kind() == reflect.String {
				thumbnailUrl = urlField.String()
			}
		} else if previewField.Kind() == reflect.String {
			// Check if Preview is directly a string URL
			thumbnailUrl = previewField.String()
		}
	}

	// If no Preview field found, try PreviewURL field
	if thumbnailUrl == "" {
		previewURLField := v.FieldByName("PreviewURL")
		if previewURLField.IsValid() && previewURLField.Kind() == reflect.String {
			thumbnailUrl = previewURLField.String()
		}
	}

	if thumbnailUrl == "" {
		logger.Errorf("No Preview field found in video object for: %s", videoUrl)
		return
	}

	// Ensure thumbnail URL is valid
	if !strings.HasPrefix(thumbnailUrl, "http") {
		logger.Errorf("Invalid thumbnail URL: %s", thumbnailUrl)
		return
	}

	logger.Infof("Trying to download thumbnail from: %s", thumbnailUrl)

	// Generate thumbnail filename (same as video but with .jpg extension)
	segments := strings.Split(videoUrl, "/")
	videoName := segments[len(segments)-1]
	re := regexp.MustCompile(`name=`)
	if re.MatchString(videoName) {
		segments := strings.Split(videoName, "?")
		videoName = segments[len(segments)-2]
	}

	// Get tweet content for filename
	tweetContent := "没有推文"
	pattern := `[/\\:*?"<>|]`
	regex, _ := regexp.Compile(pattern)

	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
		}
	case *twitterscraper.Tweet:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
		}
	}

	// Create thumbnail filename
	nameWithoutExt := strings.TrimSuffix(videoName, "."+strings.Split(videoName, ".")[len(strings.Split(videoName, "."))-1])
	thumbnailName := nameWithoutExt + "_" + tweetContent + ".jpg"

	// Download thumbnail
	req, err := http.NewRequest("GET", thumbnailUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Errorf("Error downloading thumbnail: %s", err.Error())
		return
	}
	if resp.StatusCode != 200 {
		logger.Errorf("Error downloading thumbnail, status code: %d", resp.StatusCode)
		return
	}

	logger.Infof("Thumbnail download started")

	// Save thumbnail to video directory
	var f *os.File
	if dwn_type == "user" {
		if _, err := os.Stat(output + "/video"); os.IsNotExist(err) {
			os.MkdirAll(output+"/video", os.ModePerm)
		}
		f, _ = os.Create(output + "/video/" + thumbnailName)
	} else {
		if _, err := os.Stat(output); os.IsNotExist(err) {
			os.MkdirAll(output, os.ModePerm)
		}
		f, _ = os.Create(output + "/" + thumbnailName)
	}
	if f != nil {
		defer f.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			logger.Errorf("Failed to save thumbnail: %s", err.Error())
			return
		}
		logger.Infof("Downloaded thumbnail: %s", thumbnailName)
	}
}

func download(wg *sync.WaitGroup, tweet interface{}, url string, filetype string, output string, dwn_type string) {
	defer wg.Done()
	segments := strings.Split(url, "/")
	name := segments[len(segments)-1]
	re := regexp.MustCompile(`name=`)
	if re.MatchString(name) {
		segments := strings.Split(name, "?")
		name = segments[len(segments)-2]
	}

	// Log download start
	logger.Infof("Starting download: %s", name)
	logger.Infof("URL: %s", url)
	logger.Infof("File type: %s", filetype)
	logger.Infof("Output directory: %s", output)
	// Get tweet content
	tweetContent := "没有推文"
	pattern := `[/\\:*?"<>|]`
	regex, _ := regexp.Compile(pattern)

	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
		}
	case *twitterscraper.Tweet:
		if t.Text != "" {
			tweetContent = sanitizeText(t.Text, regex, 240)
		}
	}

	// Add tweet content to filename
	nameWithoutExt := strings.TrimSuffix(name, "."+strings.Split(name, ".")[len(strings.Split(name, "."))-1])
	ext := "." + strings.Split(name, ".")[len(strings.Split(name, "."))-1]
	name = nameWithoutExt + "_" + tweetContent + ext

	if format != "" {
		name = getFormat(tweet) + "_" + name
	}
	if urlOnly {
		logger.Info(url)
		time.Sleep(2 * time.Millisecond)
		return
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Errorf("Download failed: %s", err.Error())
		return
	}

	if resp.StatusCode != 200 {
		logger.Errorf("Download failed with status code: %d", resp.StatusCode)
		return
	}

	logger.Infof("Download started: %s", name)

	var f *os.File
	if dwn_type == "user" {
		if update {
			if _, err := os.Stat(output + "/" + filetype + "/" + name); !errors.Is(err, os.ErrNotExist) {
				logger.Infof("File already exists: %s", name)
				return
			}
		}
		if filetype == "rtimg" {
			f, _ = os.Create(output + "/img/RE-" + name)
		} else if filetype == "rtvideo" {
			f, _ = os.Create(output + "/video/RE-" + name)
		} else {
			f, _ = os.Create(output + "/" + filetype + "/" + name)
		}
	} else {
		if update {
			if _, err := os.Stat(output + "/" + name); !errors.Is(err, os.ErrNotExist) {
				logger.Infof("File already exists: %s", name)
				return
			}
		}
		f, _ = os.Create(output + "/" + name)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		logger.Errorf("Failed to save file: %s", err.Error())
		return
	}
	logger.Infof("Download completed: %s", name)
}

func videoUser(wait *sync.WaitGroup, tweet *twitterscraper.TweetResult, output string, rt bool) {
	defer wait.Done()
	wg := sync.WaitGroup{}
	if len(tweet.Videos) > 0 {
		logger.Infof("Processing %d videos for tweet: %s", len(tweet.Videos), tweet.ID)
		for _, i := range tweet.Videos {
			url := strings.Split(i.URL, "?")[0]
			logger.Infof("Processing video: %s", url)
			if tweet.IsRetweet {
				if rt || onlyrtw {
					wg.Add(1)
					go download(&wg, tweet, url, "video", output, "user")
					// Download video thumbnail
					wg.Add(1)
					go downloadThumbnail(&wg, tweet, i, url, output, "user")
					// Generate NFO file
					generateNFOFile(tweet, url, output, "user")
					// Generate ASS subtitle file
					generateASSFile(tweet, url, output, "user")
					continue
				} else {
					continue
				}
			} else if onlyrtw {
				continue
			}
			wg.Add(1)
			go download(&wg, tweet, url, "video", output, "user")
			// Download video thumbnail
			wg.Add(1)
			go downloadThumbnail(&wg, tweet, i, url, output, "user")
			// Generate NFO file
			generateNFOFile(tweet, url, output, "user")
			// Generate ASS subtitle file
			generateASSFile(tweet, url, output, "user")
		}
		wg.Wait()
	}
}

func photoUser(wait *sync.WaitGroup, tweet *twitterscraper.TweetResult, output string, rt bool) {
	defer wait.Done()
	wg := sync.WaitGroup{}
	if len(tweet.Photos) > 0 || tweet.IsRetweet {
		if tweet.IsRetweet && (rt || onlyrtw) {
			singleTweet(output, tweet.ID)
		}
		for _, i := range tweet.Photos {
			if onlyrtw || tweet.IsRetweet {
				continue
			}
			var url string
			if !strings.Contains(i.URL, "video_thumb/") {
				if size == "orig" || size == "small" {
					url = i.URL + "?name=" + size
				} else {
					url = i.URL
				}
				wg.Add(1)
				go download(&wg, tweet, url, "img", output, "user")
			}
		}
		wg.Wait()
	}
}

func videoSingle(tweet *twitterscraper.Tweet, output string) {
	if tweet == nil {
		return
	}
	if len(tweet.Videos) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.Videos {
			url := strings.Split(i.URL, "?")[0]
			if usr != "" {
				wg.Add(1)
				go download(&wg, tweet, url, "rtvideo", output, "user")
				// Download video thumbnail
				wg.Add(1)
				go downloadThumbnail(&wg, tweet, i, url, output, "user")
				// Generate NFO file
				generateNFOFile(tweet, url, output, "user")
				// Generate ASS subtitle file
				generateASSFile(tweet, url, output, "user")
			} else {
				wg.Add(1)
				go download(&wg, tweet, url, "tweet", output, "tweet")
				// Download video thumbnail
				wg.Add(1)
				go downloadThumbnail(&wg, tweet, i, url, output, "tweet")
				// Generate NFO file
				generateNFOFile(tweet, url, output, "tweet")
				// Generate ASS subtitle file
				generateASSFile(tweet, url, output, "tweet")
			}
		}
		wg.Wait()
	}
}

func photoSingle(tweet *twitterscraper.Tweet, output string) {
	if tweet == nil {
		return
	}
	if len(tweet.Photos) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.Photos {
			var url string
			if !strings.Contains(i.URL, "video_thumb/") {
				if size == "orig" || size == "small" {
					url = i.URL + "?name=" + size
				} else {
					url = i.URL
				}
				if usr != "" {
					wg.Add(1)
					go download(&wg, tweet, url, "rtimg", output, "user")
				} else {
					wg.Add(1)
					go download(&wg, tweet, url, "tweet", output, "tweet")
				}
			}
		}
		wg.Wait()
	}
}

func processCookieString(cookieStr string) []*http.Cookie {
	cookiePairs := strings.Split(cookieStr, "; ")
	cookies := make([]*http.Cookie, 0)
	expiresTime := time.Now().AddDate(1, 0, 0)

	for _, pair := range cookiePairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		value := parts[1]
		value = strings.Trim(value, "\"")

		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Path:     "/",
			Domain:   ".x.com",
			Expires:  expiresTime,
			HttpOnly: true,
			Secure:   true,
		}

		cookies = append(cookies, cookie)
	}
	return cookies
}

func askPass() {
	for {
		var auth_token, ct0 string
		logger.Info(`  ╔═══════════════════════════════════════════════════════════════╗
  ║                                                               ║
  ║  User/pass login is no longer supported,                      ║
  ║  Log in using a browser and find auth_token and ct0 cookies.  ║
  ║  (via Inspect → Storage → Cookies).                           ║
  ║                                                               ║
  ╚═══════════════════════════════════════════════════════════════╝`)
		logger.Info()
		fmt.Printf("auth_token cookie: ")
		fmt.Scanln(&auth_token)
		fmt.Printf("ct0 cookie: ")
		fmt.Scanln(&ct0)
		scraper.SetAuthToken(twitterscraper.AuthToken{Token: auth_token, CSRFToken: ct0})
		if !scraper.IsLoggedIn() {
			logger.Error("Bad Cookies.")
			askPass()
		}
		cookies := scraper.GetCookies()
		js, _ := json.Marshal(cookies)
		f, _ := os.OpenFile("twmd_cookies.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		defer f.Close()
		f.Write(js)
		break
	}
}

func Login(useCookies bool) {
	if useCookies {
		if _, err := os.Stat("twmd_cookies.json"); errors.Is(err, fs.ErrNotExist) {
			logger.Info("Enter cookies string: ")
			var cookieStr string
			cookieStr, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			cookieStr = strings.TrimSpace(cookieStr)

			cookies := processCookieString(cookieStr)
			scraper.SetCookies(cookies)

			// Save cookies to file
			js, _ := json.MarshalIndent(cookies, "", "  ")
			f, _ := os.OpenFile("twmd_cookies.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			defer f.Close()
			f.Write(js)
		} else {
			f, _ := os.Open("twmd_cookies.json")
			var cookies []*http.Cookie
			json.NewDecoder(f).Decode(&cookies)
			scraper.SetCookies(cookies)
			logger.Info(scraper.IsLoggedIn())
		}
	} else {
		if _, err := os.Stat("twmd_cookies.json"); errors.Is(err, fs.ErrNotExist) {
			askPass()
		} else {
			f, _ := os.Open("twmd_cookies.json")
			var cookies []*http.Cookie
			json.NewDecoder(f).Decode(&cookies)
			scraper.SetCookies(cookies)
		}
	}

	if !scraper.IsLoggedIn() {
		if useCookies {
			logger.Error("Invalid cookies. Please try again.")
			os.Remove("twmd_cookies.json")
			Login(useCookies)
		} else {
			askPass()
		}
	} else {
		logger.Info("Logged in.")
	}
}

func singleTweet(output string, id string) {
	tweet, err := scraper.GetTweet(id)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if tweet == nil {
		logger.Error("Error retrieve tweet")
		return
	}
	if usr != "" {
		if vidz {
			videoSingle(tweet, output)
		}
		if imgs {
			photoSingle(tweet, output)
		}
	} else {
		videoSingle(tweet, output)
		photoSingle(tweet, output)
	}
}

func getFormat(tweet interface{}) string {
	var formatNew string
	var tweetResult *twitterscraper.TweetResult
	var tweetObj *twitterscraper.Tweet

	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		tweetResult = t
	case *twitterscraper.Tweet:
		tweetObj = t
	default:
		logger.Error("Invalid tweet type")
		return ""
	}

	pattern := `[/\\:*?"<>|]`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		logger.Error("Error compiling regular expression:", err)
		return ""
	}

	replacer := map[string]string{}

	if tweetResult != nil {
		replacer["{DATE}"] = time.Unix(tweetResult.Timestamp, 0).Format(datefmt)
		replacer["{NAME}"] = tweetResult.Name
		replacer["{USERNAME}"] = tweetResult.Username
		replacer["{TITLE}"] = sanitizeText(tweetResult.Text, regex, 255)
		replacer["{ID}"] = tweetResult.ID
	} else if tweetObj != nil {
		replacer["{DATE}"] = time.Unix(tweetObj.Timestamp, 0).Format(datefmt)
		replacer["{NAME}"] = tweetObj.Name
		replacer["{USERNAME}"] = tweetObj.Username
		replacer["{TITLE}"] = sanitizeText(tweetObj.Text, regex, 255)
		replacer["{ID}"] = tweetObj.ID
	}

	formatNew = format

	for key, val := range replacer {
		formatNew = strings.ReplaceAll(formatNew, key, val)
	}

	return formatNew
}

func sanitizeText(text string, regex *regexp.Regexp, maxLen int) string {
	// 1. 剔除URL
	urlRegex := regexp.MustCompile(`https?://[\w\-._~:/?#[\]@!$&'()*+,;=.]+`)
	text = urlRegex.ReplaceAllString(text, "")

	// 2. 剔除换行符和制表符
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// 3. 剔除连续空格
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// 4. 剔除emoji
	emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F1E0}-\x{1F1FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]`)
	text = emojiRegex.ReplaceAllString(text, "")

	// 5. 清理剩余特殊字符并限制长度
	cleaned := ""
	remaining := maxLen
	for _, char := range text {
		charStr := string(char)
		if regex.MatchString(charStr) {
			charStr = "_"
		}
		if utf8.RuneCountInString(cleaned)+utf8.RuneCountInString(charStr) > remaining {
			break
		}
		cleaned += charStr
	}

	// 6. 去除首尾空格
	cleaned = strings.TrimSpace(cleaned)

	// 7. 如果为空，返回默认值
	if cleaned == "" {
		return "无内容"
	}

	return cleaned
}

func main() {
	var nbr, single, output string
	var retweet, all, printversion, nologo, login, useCookies bool
	op := optionparser.NewOptionParser()
	op.Banner = "twmd: Apiless twitter media downloader\n\nUsage:"
	op.On("-u", "--user USERNAME", "User you want to download", &usr)
	op.On("-t", "--tweet TWEET_ID", "Single tweet to download", &single)
	op.On("-n", "--nbr NBR", "Number of tweets to download", &nbr)
	op.On("-i", "--img", "Download images only", &imgs)
	op.On("-v", "--video", "Download videos only", &vidz)
	op.On("-a", "--all", "Download images and videos", &all)
	op.On("-r", "--retweet", "Download retweet too", &retweet)
	op.On("-z", "--url", "Print media url without download it", &urlOnly)
	op.On("-R", "--retweet-only", "Download only retweet", &onlyrtw)
	op.On("-M", "--mediatweet-only", "Download only media tweet", &onlymtw)
	op.On("-s", "--size SIZE", "Choose size between small|normal|large (default large)", &size)
	op.On("-U", "--update", "Download missing tweet only", &update)
	op.On("-o", "--output DIR", "Output directory", &output)
	op.On("-f", "--file-format FORMAT", "Formatted name for the downloaded file, {DATE} {USERNAME} {NAME} {TITLE} {ID}", &format)
	op.On("-d", "--date-format FORMAT", "Apply custom date format. (https://go.dev/src/time/format.go)", &datefmt)
	op.On("-L", "--login", "Login (needed for NSFW tweets)", &login)
	op.On("-C", "--cookies", "Use cookies for authentication", &useCookies)
	op.On("-p", "--proxy PROXY", "Use proxy (proto://ip:port)", &proxy)
	op.On("-V", "--version", "Print version and exit", &printversion)
	op.On("-B", "--no-banner", "Don't print banner", &nologo)
	op.Exemple("twmd -u Spraytrains -o ~/Downloads -a -r -n 300")
	op.Exemple("twmd -u Spraytrains -o ~/Downloads -R -U -n 300")
	op.Exemple("twmd --proxy socks5://127.0.0.1:9050 -t 156170319961391104")
	op.Exemple("twmd -t 156170319961391104")
	op.Exemple("twmd -t 156170319961391104 -f \"{DATE} {ID}\"")
	op.Exemple("twmd -t 156170319961391104 -f \"{DATE} {ID}\" -d \"2006-01-02_15-04-05\"")
	op.Parse()

	if printversion {
		logger.Info("version:", version)
		os.Exit(1)
	}

	op.Logo("twmd", "elite", nologo)
	if usr == "" && single == "" {
		logger.Error("You must specify an user (-u --user) or a tweet (-t --tweet)")
		op.Help()
		os.Exit(1)
	}
	if all {
		vidz = true
		imgs = true
	}
	if !vidz && !imgs && single == "" {
		logger.Error("You must specify what to download. (-i --img) for images, (-v --video) for videos or (-a --all) for both")
		op.Help()
		os.Exit(1)
	}
	var re = regexp.MustCompile(`{ID}|{DATE}|{NAME}|{USERNAME}|{TITLE}`)
	if format != "" && !re.MatchString(format) {
		logger.Error("You must specify a format (-f --format)")
		op.Help()
		os.Exit(1)
	}

	re = regexp.MustCompile("small|normal|large")
	if !re.MatchString(size) && size != "orig" {
		logger.Error("Error in size, setting up to normal")
		size = ""
	}
	if size == "large" {
		size = "orig"
	}

	client = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Duration(5) * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   time.Duration(5) * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
		},
	}
	if proxy != "" {
		proxyURL, _ := URL.Parse(proxy)
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	scraper = twitterscraper.New()
	scraper.WithReplies(true)
	scraper.SetProxy(proxy)

	// Modified login handling
	if login || useCookies {
		Login(useCookies)
	}

	if single != "" {
		if output == "" {
			output = "./"
		} else {
			os.MkdirAll(output, os.ModePerm)
		}
		singleTweet(output, single)
		os.Exit(0)
	}
	if nbr == "" {
		nbr = "3000"
	}
	if output != "" {
		output = output + "/" + usr
	} else {
		output = usr
	}
	if vidz {
		os.MkdirAll(output+"/video", os.ModePerm)
	}
	if imgs {
		os.MkdirAll(output+"/img", os.ModePerm)
	}
	nbrs, _ := strconv.Atoi(nbr)
	wg := sync.WaitGroup{}

	var tweets <-chan *twitterscraper.TweetResult
	if onlymtw {
		tweets = scraper.GetMediaTweets(context.Background(), usr, nbrs)
	} else {
		tweets = scraper.GetTweets(context.Background(), usr, nbrs)
	}

	for tweet := range tweets {
		if tweet.Error != nil {
			logger.Error(tweet.Error)
			os.Exit(1)
		}
		if vidz {
			wg.Add(1)
			go videoUser(&wg, tweet, output, retweet)
		}
		if imgs {
			wg.Add(1)
			go photoUser(&wg, tweet, output, retweet)
		}
	}
	wg.Wait()
}
