const childProcess = require('child_process');
const getShell = require('./get-shell');

function spawn(command, spawnOpts) {
  const shellDetails = getShell();

  return childProcess.spawn(
    shellDetails.shell,
    [shellDetails.executeFlag, command],
    spawnOpts
  );
}

module.exports = {
  spawn: spawn,
};
