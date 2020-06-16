package chrometool

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"testing"

	"github.com/chromedp/chromedp"
)

var flagPort = flag.Int("port", 8544, "port")

func TestChromeTool(t *testing.T) {
	flag.Parse()

	// run server
	go testServer(fmt.Sprintf(":%d", *flagPort))

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// run task list
	var val1, val2 string
	err := chromedp.Run(ctx, sendActions(fmt.Sprintf("http://localhost:%d", *flagPort), &val1, &val2))
	if err != nil {
		t.Errorf("Encountered error %s in Run()", err.Error())
	}

	want := "textarea2test2"
	got := val1
	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}

	want = "Click button 2"
	got = val2
	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}
}

// sendActions sends actions to the server and extracts 4 values from the html page.
func sendActions(host string, val1, val2 *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(host),
		SendKeysToNthElement(`.keystext`, 2, "test2"),
		chromedp.Value(`#testarea2`, val1, chromedp.ByID),
		ClickNthElement(`clicktest`, 2, chromedp.ByQueryAll),
		chromedp.Value(`#p2`, val2, chromedp.ByID),
	}
}

// testServer is a simple HTTP server that displays the passed headers in the html.
func testServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(res, indexHTML)
	})
	return http.ListenAndServe(addr, mux)
}

const indexHTML = `<!doctype html>
<html>
<head>
  <title>test</title>
</head>
<body>
  <div id="box1">   
    <p id="p1">
	  <textarea id="textarea1" class="keystest" style="width:500px;height:400px">textarea1</textarea><br><br>
	  <textarea id="textarea2" class="keystest" style="width:500px;height:400px">textarea2</textarea><br><br>
	  <textarea id="textarea3" class="keystest" style="width:500px;height:400px">textarea3</textarea><br><br>
    </p>
  </div>
  <div id="box2"
  	<p id='p2'>para2</p> 
	  <input type='button' class="clicktest" onclick='changeText1()' value='Change Text 1'/>
	  <input type='button' class="clicktest" onclick='changeText2()' value='Change Text 2'/>
	  <input type='button' class="clicktest" onclick='changeText3()' value='Change Text 3'/>
  </div>
<script>
function changeText1()
{
 document.getElementById('p2').innerHTML = 'Click button 1';
}
function changeText2()
{
 document.getElementById('p2').innerHTML = 'Click button 2';
}
function changeText3()
{
 document.getElementById('p2').innerHTML = 'Click button 3';
}
</script>
</body>
</html>`
