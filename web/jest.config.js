module.exports = {
    transform: {
      '^.+\\.svelte$': ['svelte-jester', { "preprocess": true }],
      '^.+\\.ts$': 'ts-jest',
    },
    testRegex: '(/__tests__/.*|\\.(test))\\.(ts|js)$',
    moduleFileExtensions: ['ts', 'js', 'svelte'],
  }