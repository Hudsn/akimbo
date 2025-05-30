window.addEventListener('load', function () {

    const eventSource = new EventSource("~~urlpath~~")

    eventSource.onmessage = function (event) {
        console.log("GOT EVENT")
        const data = JSON.parse(event.data)
        const message = data.message
        if (message === "reload") {
            window.location.reload()
        }
    }

    eventSource.onopen = function () {
        console.log("connected to sse reloader endpoint ~~urlpath~~")
    }

    eventSource.onerror = function (error) {
        console.error("SSE error:", error)
        if (eventSource.readyState === EventSource.CLOSED) {
            console.log("Connection to SSE closed.")
        }
    }
});