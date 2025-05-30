## Akimbo Reloader

### What is it?
A drop-in browser hot reloader for Go applications.

Simply use the helpers in this package to generate a HandlerFunc, add the script to your frontend, whether that be Go templates, static html, or an SPA, and enjoy automatic browser reloading on save!

### How does it work?
When target files on your system are changed, the Akimbo handler sends a 
signal to the browser script to trigger a browser reload.

## How do I use it?

### Installation

```
go get github.com/hudsn/akimbo
```

### Usage

#### Create the config:
```go
// create the config
config := akimbo.Config{
    // the http endpoint that communicates with the browser script
    UrlPath: "/my_reloader", 

    //the file types that you want to trigger a change.
    // Note that an empty list will reload on ANY file change.
    Extensions: []string{"css", "js", "html"}, 

    // the folders containing files that you want to watch for changes(includes all subfolders). 
    // Note that this also works with "." to include the entire project.
    // Note that an empty path will default to single entry of "." (reloads on any project file change)
    Paths []string{"example/static"}, 
}
```

#### Create the reloader object:

```go
reloader, err := akimbo.NewReloader(reloadConfig)
if err != nil {
    log.Fatal(err)
}
```

#### Add the endpoints to your router/muxer/whatever you want to call it:

```go

// The context passed here should be global for your app.
// See the example/server.go file for an example on how global context is passed to this handlerfunc
router.HandleFunc(reloader.SSEHandlerPath(), reloader.SSEHandler(ctx))

// passing "true" here, will print the script tag on server startup for you to copy. This only needs to be done once as long as your config doesn't change.
router.HandleFunc(reloader.ScriptHandlerPath(), reloader.ScriptHandler(true))
```

#### Copy the script tag into your root html

```html
...
<head>
    ...
    <script defer src="/my_reloader_script"></script>
</head>
...
```

Refresh the page manually once to ensure the script tag is loaded in your current page.

#### Thats it!

Any page that includes this script tag should automatically reload the browser when one of the files that you're watching changes.

#### Note
You can also access the script tag as a string via `myreloader.ScriptTag()` if you want to dynamically add it via someting like Go HTML Templates or Templ. 

#### Warning
Be aware that any **embedded** templates or files won't change on **browser** reload since they're compiled into the temporary binary at build time, and won't reflect source file changes.

To circumvent this, when developing you can serve files directly from your local system, and then when building a project you can use the embedded version of those files.