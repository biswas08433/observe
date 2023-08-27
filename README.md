# Observe
Concurrent monitoring Program for executables.


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

    //Add the folders to observe
    "folders": ["./Public","./Views"]
}
```
