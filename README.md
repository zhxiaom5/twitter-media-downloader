# twmd: CLI twitter media downloader (without api key)

This twitter downloader doesn't require Credentials or an api key. It's based on [twitter-scrapper](https://github.com/imperatrona/twitter-scraper).

## Recent Changes

### v1.15.0
- **Filename Generation Improvement**: 
  - Enhanced filename sanitization to handle special characters, URLs, and emojis
  - Adjusted tweet content length in filenames to 240 characters for better readability

- **ASS Subtitle Generation**: 
  - Added automatic ASS subtitle file generation for videos
  - Supports automatic line wrapping for long subtitles
  - Configured subtitles to be bottom-center aligned with proper margins
  - Optimized subtitle resolution settings (PlayResX: 1080, PlayResY: 1920)

- **Logging System Upgrade**: 
  - Replaced all `fmt.Printf` and `fmt.Println` calls with logrus logging methods
  - Added detailed logging for all operations
  - Configured logging format with timestamps and log levels

- **Other Improvements**: 
  - Added comprehensive error handling
  - Optimized download process
  - Improved proxy support

### Note
For NSFW or private accounts, you will need to logged in (-L). Username ans password login isn't supported anymore. You'll have to logged in in a browser and copy auth_token and ct0 cookies (right click => inspect => storage => cookies).
It will create a twmd_cookies.json so you will not have to enter these cookies everytime.


![gui](.github/screenshots/gui.png)
**Note:** Gui is not longer maintained.

## usage: 

```
Usage:
-h, --help                   Show this help
-u, --user=USERNAME          User you want to download
-t, --tweet=TWEET_ID         Single tweet to download
-n, --nbr=NBR                Number of tweets to download
-i, --img                    Download images only
-v, --video                  Download videos only
-a, --all                    Download images and videos
-r, --retweet                Download retweet too
-z, --url                    Print media url without download it
-R, --retweet-only           Download only retweet
-M, --mediatweet-only        Download only media tweet
-s, --size=SIZE              Choose size between small|normal|large (default
                             large)
-U, --update                 Download missing tweet only
-o, --output=DIR             Output directory
-f, --file-format=FORMAT     Formatted name for the downloaded file, {DATE}
                             {USERNAME} {NAME} {TITLE} {ID}
-d, --date-format=FORMAT     Apply custom date format.
                             (https://go.dev/src/time/format.go)
-L, --login                  Login (needed for NSFW tweets)
-C, --cookies                Use cookies for authentication
-p, --proxy=PROXY            Use proxy (proto://ip:port)
-V, --version                Print version and exit
-B, --no-banner              Don't print banner
```

### Examples:

#### Download 300 tweets from @Spraytrains.

If the tweet doesn't contain a photo or video nothing will be downloaded but it will count towards the 300.

```sh
twmd -u Spraytrains -o ~/Downloads -a -n 300
```

Due to rate limits of twitter, it is possible to fetch at most 500–600 tweets.
To fetch as more tweets as possible, change the argument of `-n` to a bigger number, like 3000.

You can use `-r|--retweet` to download retweets as well, or `-R|--retweet-only` to download retweet only

`-U|--update` will only download missing media.

#### Download a single tweet:

```sh
twmd -t 156170319961391104
```

#### NSFW tweets

You'll need to login `-L|--login` for downloading nsfw tweets. Or you can provide cookies `-C|--cookies` to complete the login.


#### Using proxy

Both http and socks4/5 can be used:

```sh
twmd  --proxy socks5://127.0.0.1:9050 -t 156170319961391104
```

### Installation:


**Note:** If you don't want to build it you can download prebuilt binaries [here](https://github.com/mmpx12/twitter-media-downloader/releases/latest).


#### Cli:

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
make
sudo make install
# OR
sudo make all
# Clean
sudo make clean
```

#### Gui (outdated):

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
# linux
make linux-gui
# windows
make windows-gui
```


#### Termux (no root):

installation: 

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
make
make termux-install
# OR
make termux-all
# Clean
make termux-clean
```

You may also want to add stuff in ~/bin/termux-url-opener to automatically download profile or post when share with termux.

```sh
cd ~/storage/downlaods
if grep twitter <<< "$1" >/dev/null; then
  if [[ $(tr -cd '/' <<< "$1" | wc -c) -eq 3 ]]; then
    userid=$(cut -d '/' -f 4 <<< "$1" |  cut -d '?' -f 1)
    echo "$userid"
    twmd -B -u "$userid" -o twitter -i -v -n 3000
  else 
    postid=$(cut -d '/' -f 6 <<< "$1" |  cut -d '?' -f 1)
    twmd -B -t "$postid" -o twitter
  fi
fi
```


Check [here](https://gist.github.com/mmpx12/f0741d40909ed3f182fd6f9b33b580d7) for a full termux-url-opener example.


#### Gifs are not supported at the moment.

---

# twmd: CLI Twitter 媒体下载器（无需 API 密钥）

这个 Twitter 下载器不需要凭证或 API 密钥。它基于 [twitter-scrapper](https://github.com/imperatrona/twitter-scraper) 构建。

## 最近变更

### v1.15.0
- **文件名生成改进**：
  - 增强了文件名清理功能，可处理特殊字符、URL 和表情符号
  - 将文件名中的推文内容长度调整为 240 个字符，以提高可读性

- **ASS 字幕生成**：
  - 为视频添加了自动 ASS 字幕文件生成
  - 支持长字幕的自动换行
  - 配置了字幕位置为底部居中，带有适当的边距
  - 优化了字幕分辨率设置（PlayResX: 1080, PlayResY: 1920）

- **日志系统升级**：
  - 将所有 `fmt.Printf` 和 `fmt.Println` 调用替换为 logrus 日志方法
  - 为所有操作添加了详细的日志记录
  - 配置了带有时间戳和日志级别的日志格式

- **其他改进**：
  - 添加了全面的错误处理
  - 优化了下载过程
  - 改进了代理支持

### 注意
对于 NSFW 或私人账户，您需要登录 (-L)。用户名和密码登录不再受支持。您需要在浏览器中登录并复制 auth_token 和 ct0 cookies（右键 => 检查 => 存储 => cookies）。
它会创建一个 twmd_cookies.json 文件，因此您不必每次都输入这些 cookies。


![gui](.github/screenshots/gui.png)
**注意：** GUI 不再维护。

## 使用方法：

```
用法：
-h, --help                   显示此帮助
-u, --user=USERNAME          要下载的用户
-t, --tweet=TWEET_ID         要下载的单个推文
-n, --nbr=NBR                要下载的推文数量
-i, --img                    仅下载图片
-v, --video                  仅下载视频
-a, --all                    下载图片和视频
-r, --retweet                也下载转推
-z, --url                    打印媒体 URL 而不下载
-R, --retweet-only           仅下载转推
-M, --mediatweet-only        仅下载媒体推文
-s, --size=SIZE              选择大小：small|normal|large（默认 large）
-U, --update                 仅下载缺失的媒体
-o, --output=DIR             输出目录
-f, --file-format=FORMAT     下载文件的格式化名称，{DATE} {USERNAME} {NAME} {TITLE} {ID}
-d, --date-format=FORMAT     应用自定义日期格式。
                             (https://go.dev/src/time/format.go)
-L, --login                  登录（NSFW 推文需要）
-C, --cookies                使用 cookies 进行身份验证
-p, --proxy=PROXY            使用代理（proto://ip:port）
-V, --version                打印版本并退出
-B, --no-banner              不打印横幅
```

### 示例：

#### 从 @Spraytrains 下载 300 条推文。

如果推文不包含照片或视频，则不会下载任何内容，但会计入 300 条。

```sh
twmd -u Spraytrains -o ~/Downloads -a -n 300
```

由于 Twitter 的速率限制，最多可以获取 500-600 条推文。
要获取尽可能多的推文，请将 `-n` 的参数更改为更大的数字，如 3000。

您可以使用 `-r|--retweet` 来下载转推，或使用 `-R|--retweet-only` 仅下载转推

`-U|--update` 将仅下载缺失的媒体。

#### 下载单个推文：

```sh
twmd -t 156170319961391104
```

#### NSFW 推文

您需要登录 `-L|--login` 才能下载 NSFW 推文。或者您可以提供 cookies `-C|--cookies` 来完成登录。


#### 使用代理

支持 http 和 socks4/5：

```sh
twmd  --proxy socks5://127.0.0.1:9050 -t 156170319961391104
```

### 安装：


**注意：** 如果您不想构建它，您可以在此处下载预构建的二进制文件 [here](https://github.com/mmpx12/twitter-media-downloader/releases/latest)。


#### 命令行：

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
make
sudo make install
# 或
sudo make all
# 清理
sudo make clean
```

#### 图形界面（已过时）：

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
# linux
make linux-gui
# windows
make windows-gui
```


#### Termux（无需 root）：

安装：

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
make
make termux-install
# 或
make termux-all
# 清理
make termux-clean
```

您可能还想在 ~/bin/termux-url-opener 中添加内容，以便在与 Termux 共享时自动下载个人资料或帖子。

```sh
cd ~/storage/downlaods
if grep twitter <<< "$1" >/dev/null; then
  if [[ $(tr -cd '/' <<< "$1" | wc -c) -eq 3 ]]; then
    userid=$(cut -d '/' -f 4 <<< "$1" |  cut -d '?' -f 1)
    echo "$userid"
    twmd -B -u "$userid" -o twitter -i -v -n 3000
  else 
    postid=$(cut -d '/' -f 6 <<< "$1" |  cut -d '?' -f 1)
    twmd -B -t "$postid" -o twitter
  fi
fi
```


请在此处查看完整的 termux-url-opener 示例 [here](https://gist.github.com/mmpx12/f0741d40909ed3f182fd6f9b33b580d7)。


#### 目前不支持 GIF。
