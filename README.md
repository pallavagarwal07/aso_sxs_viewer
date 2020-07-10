# Search Overlay Side by Side Viewer
The viewer allows you to compare search-as-you-type results from multiple accounts/URLs side by side by querying them simultaneously. Assisting in comparing the results visually.

<img src="/static/googleandyoutubesearch.png">



## Installation
### Linux

### MacOS

## Usage
You will find config.textproto in the $HOME/.aso_sxs_viewer/ 

```
browser_window_count:  2
url:  "https://mail.google.com/"
css_selector:  {
	selector:  "input"
	position:  7
}
user_data_dir_prefix:  "$HOME/.aso_sxs_viewer/profiles"
input_window_position:  DEFAULT_BOTTOM
root_window_config:  {
	layout:  {
		width:  1600
		height:  900
	}
}
# You may use the template below to add window_overrides	
#	window_overrides: {
#		url: "https://mail.google.com/"
#		css_selector: {
#			selector: "input"
#			position: 7
#		}
#	}
```

## Development

install bazel [bazel installation](https://docs.bazel.build/versions/master/install.html)

