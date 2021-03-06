import svelte from 'rollup-plugin-svelte';
import commonjs from '@rollup/plugin-commonjs';
import resolve from '@rollup/plugin-node-resolve';
import livereload from 'rollup-plugin-livereload';
import { terser } from 'rollup-plugin-terser';
import sveltePreprocess from 'svelte-preprocess';
import typescript from '@rollup/plugin-typescript';
import replace from '@rollup/plugin-replace';
import workerLoader from 'rollup-plugin-web-worker-loader';
import postcss from 'rollup-plugin-postcss';
import rollupJson from '@rollup/plugin-json';

const production = !process.env.ROLLUP_WATCH;
const AUTH0_CLIENTID = process.env.AUTH0_CLIENTID;
const AUTH0_DOMAIN = process.env.AUTH0_DOMAIN;
const API_URL = process.env.API_URL;
const TRANSAK_API_KEY = process.env.TRANSAK_API_KEY;
const TRANSAK_ENV = process.env.TRANSAK_ENV;

function serve() {
	let server;

	function toExit() {
		if (server) server.kill(0);
	}

	return {
		writeBundle() {
			if (server) return;
			server = require('child_process').spawn('npm', ['run', 'start', '--', '--dev'], {
				stdio: ['ignore', 'inherit', 'inherit'],
				shell: true
			});

			process.on('SIGTERM', toExit);
			process.on('exit', toExit);
		}
	};
}

export default {
	input: 'src/main.ts',
	output: {
		sourcemap: true,
		format: 'es',
		name: 'app',
		dir: 'public/build/'
	},
	plugins: [
		replace({
			preventAssignment: true,
			'process.env.AUTH0_CLIENTID': JSON.stringify(AUTH0_CLIENTID ? AUTH0_CLIENTID : "DcMwCcm9VNE3xMz6Sxtde8FqdXH8Berq"),
			'process.env.AUTH0_DOMAIN': JSON.stringify(AUTH0_DOMAIN ? AUTH0_DOMAIN : "https://dev-xfscxtiv.us.auth0.com"),
			'process.env.API_URL': JSON.stringify(API_URL ? API_URL : "http://localhost:8080/api"),
      'process.env.TRANSAK_API_KEY': JSON.stringify(TRANSAK_API_KEY ? TRANSAK_API_KEY : "bec7a499-5b22-4928-832e-8abd938305b5"),
      'process.env.TRANSAK_ENV': JSON.stringify(TRANSAK_ENV ? TRANSAK_ENV : "STAGING"),
		}),
    workerLoader({
      targetPlatform: 'browser',
      inline: 'false',
      outputFolder: 'public',
    }),
		svelte({
			preprocess: sveltePreprocess({ sourceMap: !production }),
			compilerOptions: {
				// enable run-time checks when not in production
				dev: !production
			}
		}),
		
    // we'll extract any component CSS out into
		// a separate file - better for performance
    postcss({
      extract: 'bundle.css',
      minimize: production,
      use: [
        [
          'sass',
          {
            includePaths: ['./src/theme/dark', './node_modules'],
          },
        ],
      ],
    }),

		// If you have external dependencies installed from
		// npm, you'll most likely need these plugins. In
		// some cases you'll need additional configuration -
		// consult the documentation for details:
		// https://github.com/rollup/plugins/tree/master/packages/commonjs
		resolve({
			browser: true,
			dedupe: ['svelte']
		}),
    rollupJson(),
		commonjs(),
		typescript({
			sourceMap: !production,
			inlineSources: !production
		}),

		// In dev mode, call `npm run start` once
		// the bundle has been generated
		!production && serve(),

		// Watch the `public` directory and refresh the
		// browser on changes when not in production
		!production && livereload('public'),

		// If we're building for production (npm run build
		// instead of npm run dev), minify
		production && terser()
	],
	watch: {
		clearScreen: false
	}
};
