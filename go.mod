module github.com/go-spook/spook

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/alecthomas/chroma v0.6.6
	github.com/fatih/color v1.7.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/tdewolff/minify v2.3.6+incompatible
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tdewolff/test v1.0.3 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.1
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
