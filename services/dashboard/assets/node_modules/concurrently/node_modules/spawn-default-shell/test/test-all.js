const _ = require('lodash');
const testBasic = require('./test-basic');
const testPlatformShared = require('./test-platform-shared');

const PLATFORMS = ['darwin', 'freebsd', 'linux', 'sunos', 'win32'];
const originalPlatform = Object.getOwnPropertyDescriptor(process, 'platform');

describe('spawn-default-shell', () => {
  testBasic();
});

describe('shared tests on each platform (mocking)', () => {
  _.each(PLATFORMS, (platform) => {
    describe(`process.platform = "${platform}"`, () => {
      before(() => {
        Object.defineProperty(process, 'platform', { value: platform });
      });

      after(() => {
        Object.defineProperty(process, 'platform', originalPlatform);
      });

      testPlatformShared();
    });
  });
});
