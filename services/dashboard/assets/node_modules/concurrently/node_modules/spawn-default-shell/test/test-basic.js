const assert = require('assert');
const defaultShell = require('../src/index');
const getShell = require('../src/get-shell');
const withEnv = require('./utils').withEnv;

const originalPlatform = Object.getOwnPropertyDescriptor(process, 'platform');

function testBasic() {
  it('piping should work', (done) => {
    const child = defaultShell.spawn('cat test/data/test.txt | grep 1', {
      stdio: 'pipe',
    });

    child.stdout.on('data', (data) => {
      assert.strictEqual(data.toString('utf8'), '1 äö☃\n');
    });

    child.on('close', (code) => {
      assert.strictEqual(code, 0);
      done();
    });
  });

  it('&& operator should work', (done) => {
    const child = defaultShell.spawn('echo 1 && node -e "process.exit(42)"');

    child.on('close', (code) => {
      assert.strictEqual(code, 42);
      done();
    });
  });

  describe('process.platform = "darwin"', () => {
    before(() => {
      Object.defineProperty(process, 'platform', { value: 'darwin' });
    });

    after(() => {
      Object.defineProperty(process, 'platform', originalPlatform);
    });

    it('shell resolution order should be 1. SHELL 2. /bin/bash', () => {
      withEnv({ SHELL: '' }, () => {
        assert.strictEqual(getShell().shell, '/bin/bash');
        assert.strictEqual(getShell().executeFlag, '-c');
      });

      withEnv({ SHELL: 'zsh' }, () => {
        assert.strictEqual(getShell().shell, 'zsh');
        assert.strictEqual(getShell().executeFlag, '-c');
      });
    });
  });

  describe('process.platform = "win32"', () => {
    before(() => {
      Object.defineProperty(process, 'platform', { value: 'win32' });
    });

    after(() => {
      Object.defineProperty(process, 'platform', originalPlatform);
    });

    it('shell resolution order should be 1. SHELL 2. COMSPEC 3. cmd.exe', () => {
      withEnv({ SHELL: '', COMSPEC: '' }, () => {
        assert.strictEqual(getShell().shell, 'cmd.exe');
        assert.strictEqual(getShell().executeFlag, '/c');
      });

      withEnv({ SHELL: '', COMSPEC: '\\C:\\cmd.exe' }, () => {
        assert.strictEqual(getShell().shell, '\\C:\\cmd.exe');
        assert.strictEqual(getShell().executeFlag, '/c');
      });

      withEnv({ SHELL: 'bash', COMSPEC: '\\C:\\cmd.exe' }, () => {
        assert.strictEqual(getShell().shell, 'bash');
        assert.strictEqual(getShell().executeFlag, '-c');
      });
    });
  });

  describe('process.platform = "linux" (other than win32 or darwin)', () => {
    before(() => {
      Object.defineProperty(process, 'platform', { value: 'linux' });
    });

    after(() => {
      Object.defineProperty(process, 'platform', originalPlatform);
    });

    it('shell resolution order should be 1. SHELL 2. /bin/sh', () => {
      withEnv({ SHELL: '' }, () => {
        assert.strictEqual(getShell().shell, '/bin/sh');
        assert.strictEqual(getShell().executeFlag, '-c');
      });

      withEnv({ SHELL: 'zsh' }, () => {
        assert.strictEqual(getShell().shell, 'zsh');
        assert.strictEqual(getShell().executeFlag, '-c');
      });
    });
  });
}

module.exports = testBasic;
