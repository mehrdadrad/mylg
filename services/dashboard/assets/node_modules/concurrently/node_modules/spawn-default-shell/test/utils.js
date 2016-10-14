const _ = require('lodash');

function withEnv(env, func) {
  const originals = _.map(env, (val, key) => ({ key: key, val: process.env[key] }));
  _.each(env, (newVal, key) => {
    process.env[key] = newVal;
  });

  try {
    func();
  } finally {
    _.each(originals, (item) => {
      if (!item.val) {
        delete process.env[item.key];
      } else {
        process.env[item.key] = item.val;
      }
    });
  }
}

module.exports = {
  withEnv: withEnv,
};
