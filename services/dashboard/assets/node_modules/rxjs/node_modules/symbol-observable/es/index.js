/* global window */
import ponyfill from './ponyfill';

var root = module;

if (typeof self !== 'undefined') {
  root = self;
} else if (typeof window !== 'undefined') {
  root = window;
} else if (typeof global !== 'undefined') {
  root = global;
} else {
  root = Function('return this')();
}

var result = ponyfill(root);
export default result;
