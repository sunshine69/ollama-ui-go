# ollama-ui-go
Simple web interface to talk to ollama API

# Why

Because I am fed up with bloatware - The default one https://github.com/open-webui/open-webui is too big, too cumbersome, buggy (upgrade to latest version breaks it due to it hit unexpected tag coming from the model)

I think it does not need to be like that. Simple, easily to change and addapt. After some hours I created this.

Probably will add the image or file upload later on when I have time

# TODO

Currently just `go run .` and access it via localhost. But it is not nice.

- Make a dockerfile and configure the port to be configurable (/) done - run it `docker run --rm -p 8081:8081 stevekieu/ollama-ui-go:20250209` and access `http://localhost:8081/static/` to open the app.
- Build the image so user can run it using docker (/)
- Authentication - (/)

More todos? you can do it!
