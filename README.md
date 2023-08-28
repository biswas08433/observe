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

Let's say you want to run build and run using two commands:
```console
$ go build .
$ ./test2 < input.txt
```

Then the buildSystem is `"go"` and the arguments are 
`"build"` and `"."`

Similarly the executable is `"./test2"` and the arguments are `"<"` and `"input.txt"`

If you want to add other folders inside your working folder for observation, add the path to the `"folders"`
list.

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
