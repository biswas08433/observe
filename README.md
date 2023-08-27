# Observe
Concurrent monitoring Program for executables.

## Prerequisite
Install golang on your system.
Link: [https://go.dev/doc/install](https://go.dev/doc/install)

## Install
```
go install github.com/biswas08433/observe
```

## Usage

To initialise the folder for observation
```
observe --init
```

This will create a file called `obsconfig.json`. Modify the value as needed.

Then for starting:
```
observe --run
```

## Example `obsconfig.json`
```json
{
    "buildSystem": "go",
    "buildArgs": [
        "build",
        "."
    ],
    "executable": "./test2",
    "args": ["<", "input.txt"],

    
    "folders": ["./Public","./Views", "/MoreFolders"]
}
```
