# revolver

A simple reloader tool written in [Go](https://golang.org).

## Build
```
go build ./cmd/revolver
```

## Test
```
go test ./...
```

## Usage

When starting revolver it looks for a file called `revolver.yml` in the current
directory. This is a simple yaml file with the configuration parameters for 
running revolver. The revolver.yml file can be checked into version control
repositories.

A full featured sample revolver.yml file looks like the following:
```
dir: "."
excludeDir: [".git", "other"]
interval: 1s
action:
  - name: "build"
    pattern: ["**/*.go", "**/*.yml"]
    exclude: ["exclude", "**/exclude*"]
    build: ["build cmd 1", "build cmd 2"]
    run: "run cmd"
  - name: "test"
    pattern: ["**/*_test.go", "**/.*"]
    build: ["build cmd 1", "build cmd 2"]
```

The configuration file aims to be quite flexible.

If a string array type contains only one element, it can be written as a simple string value:
```
dir: "."
excludeDir: ".git"
interval: 1s
action:
  - name: "build"
    pattern: "**/*.go"
    exclude: "exclude"
    build: "build cmd 1"
    run: "run cmd"
  - name: "test"
    pattern: "**/*_test.go"
    build: "build cmd 1"
```

You can omit most of the parameters, the only requirement is to have at least one 
action with a build or run command:
```
action:
  - build: ["go install"]
```

If only one action is needed, the action list can be omited and the values can be 
written in the top level of the config file:
```
dir: "src"
excludeDir: ".git"
interval: 1s
pattern: "**/*.go"
exclude: "exclude"
build: "build cmd 1"
run: "run cmd"
```

A minimal sample revolver.yml file looks like the following:
```
build: "go install"
```

Config options:

Name        | Type     | Default value 
----------- | -------- | ---------------
dir         | string   | . (current dir)
excludeDir  | []string | []
interval    | duration | 500ms
action      | []Action | []

Action options:

Name    | Type     | Default value 
--------| -------- | -------------
name    | string   | 
pattern | []string | [**/*]
exclude | []string | []
build   | []string | []
run     | string   | 

Revolver can also be used without a config file. If a build(`-b`) or run(`-r`) command line 
flag is present, it will ignore the config file and configure the application with the 
specified flags instead. It is possible to add multiple excludeDir(`-ed`), patter (`-p`),
exclude(`-e`) and build(`-b`) flags (ex: ```revolver -b "echo 1" -b "echo 2"'```).

The following flags can be used:
```
Usage of revolver:
  -b value
        Build commands
  -c string
        Path to config file (default "revolver.yml")
  -d string
        Directory to watch
  -e value
        File watch exclude patterns
  -ed value
        Excluded directories
  -i duration
        Poll interval
  -p value
        File watch patterns
  -r string
        Run command
```

### File patterns

File patterns are supported for the `pattern`, `exclude` and `excludeDir` options. 
The `pattern` options defaults to every file in every directory (`**/*`), the `exclude` 
and `excludeDir` options are empty by default.

The following special terms are supported in the patterns:

Special Terms | Meaning
------------- | -------
`*`           | matches any sequence of non-path-separators
`**`          | matches any sequence of characters, including path separators
`?`           | matches any single non-path-separator character
`[class]`     | matches any single non-path-separator character against a class of characters
`{alt1,...}`  | matches a sequence of characters if one of the comma-separated alternatives matches

Any character with a special meaning can be escaped with a backslash (`\`).

Character classes support the following:

Class      | Meaning
---------- | -------
`[abc]`    | matches any single character within the set
`[a-z]`    | matches any single character in the range
`[^class]` | matches any single character which does *not* match the class

### Actions
You can specify multiple actions with different file watch patterns that execute different commands.

### Build commands
Build commands are commands that are executed when a file changes. Build commands
are executed in order and if any of them errors out, the execution chain stops.

### Run commands
Run commands are long running processes that are started when all the build 
commands are successfully executed. They are killed and restarted every time
a file changes.

## License
Authored by [Kristóf Szabó](mailto:kristofszabo@protonmail.com) and released under the MIT license.
