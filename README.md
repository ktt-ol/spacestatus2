# Status2

Shows the space status, stores the data in a mysql database for statistics and show some nice stats. 

## Requirements 

* Go
* dep

## Install


```bash
# build binary
dep ensure
go build cmd/spaceStatus/status2.go

# create config
cp config.example.toml config.toml
vim config.toml 
```

## Run

Use the systemd service file in the `init/` folder. 

```bash
./status2 
```


## Error handling

For database errors, the application exists with an error. Mqtt errors terminate the application only on startup. 
If you use the provided service file, the application will be restarted.   