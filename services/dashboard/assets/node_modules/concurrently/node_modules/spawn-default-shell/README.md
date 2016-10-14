# spawn-default-shell

> Spawn shell command with platform default shell

[![Build Status](https://travis-ci.org/kimmobrunfeldt/spawn-default-shell.svg?branch=master)](https://travis-ci.org/kimmobrunfeldt/spawn-default-shell) [![AppVeyor Build Status](https://ci.appveyor.com/api/projects/status/github/kimmobrunfeldt/spawn-default-shell?branch=master&svg=true)](https://ci.appveyor.com/project/kimmobrunfeldt/spawn-default-shell) *master branch status*

[![NPM Badge](https://nodei.co/npm/spawn-default-shell.png?downloads=true)](https://www.npmjs.com/package/spawn-default-shell)

Like `child_process.spawn` with `shell: true` option but a bit more
convenient and customizable. You can just pass the command as a string,
and it will be executed in the platform default shell. Used in [concurrently](https://github.com/kimmobrunfeldt/concurrently).

```js
// If we are in Linux / Mac, this will work
const defaultShell = require('spawn-default-shell');
const child = defaultShell.spawn('cat src/index.js | grep function');
```

Platform | Command
---------|----------
Windows  | `cmd.exe /c "..."`. If `COMSPEC` env variable is defined, it is used as shell path.
Mac      | `/bin/bash -c "..."`
Linux    | `/bin/sh -c "..."`

You can always override the shell path by defining these two environment variables:

* `SHELL=/bin/zsh`
* `SHELL_EXECUTE_FLAG=-c`

## Install

```bash
npm install spawn-default-shell --save
```

## API

### .spawn(command, [opts])

Spawns a new process of the platform default shell using the given command.

For all options, see [child_process](https://nodejs.org/api/child_process.html#child_process_child_process_spawn_command_args_options)
documentation.

## License

MIT
