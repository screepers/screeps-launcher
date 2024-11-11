package launcher

const cliMod = `
const http = require('http')
const express = require('express')
const bodyParser = require('body-parser')
const vm = require('vm')
const util = require('util')

module.exports = config => {
	if (config.cli) {
		const app = express()
		const server = http.createServer(app)
		config.cli.connectionListener = socket => {
			server.emit('connection', socket)
		}
		app.get('/greeting', (req, res) => {
			let build = ' '
			try {
				build = 'v'+require('screeps').version
			}catch(err){}
			let text = config.cli.greeting.replace('{build}', build)
			res.write(text)
			res.end()
		})
		app.post('/cli', bodyParser.text({ type: req => true }), async (req, res) => {
			const cb = (data, isResult) => {
				res.write(data + "\n")
				if (isResult) {
					res.end()
				}
			}
			const command = req.body
			const ctx = vm.createContext(config.cli.createSandbox(cb))
			try {
				const result = await vm.runInContext(command, ctx)
				if (typeof result != 'string') {
					cb(''+util.inspect(result), true)
				} else {
					cb(''+result, true)
				}
			} catch(err) {
				cb('Error: '+(err.stack || err), true)
			}
		})
	}
}
`
