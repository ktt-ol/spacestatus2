# Status2

Shows the space status, stores the data in a mysql database for statistics and show some nice stats. 

## Requirements 

To build local:
* Go
* dep

Or only Docker.

## Install

```shell script
# build binary
dep ensure
./build.sh

# create config
cp config.example.toml config.toml
vim config.toml 
```

## Build with Docker

This script creates a docker image with proper Go build environment and uses this to build the binary. All dependencies 
and cache files are stored in the `.docker-build` folder.

```shell script
./buildWithDocker.sh
```

## Run

Use the systemd service file in the `init/` folder. 

```bash
./status2 
```


## Error handling

For database errors, the application exists with an error. Mqtt errors terminate the application only on startup. 
If you use the provided service file, the application will be restarted.   