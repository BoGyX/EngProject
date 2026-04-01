import { existsSync, readFile, statSync, writeFileSync } from 'node:fs'
import { createServer } from 'node:http'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const PORT = Number(process.env.PORT || 80)
const BACKEND_ORIGIN = process.env.BACKEND_ORIGIN || 'http://backend:8000'
const FORCE_HTTPS_REDIRECT = String(process.env.FORCE_HTTPS_REDIRECT || 'false').toLowerCase() === 'true'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const DIST_DIR = path.join(__dirname, 'dist')

const mimeTypes = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
}

function getForwardedValue(value) {
  if (!value) {
    return ''
  }
  if (Array.isArray(value)) {
    return String(value[0] || '').split(',')[0].trim()
  }
  return String(value).split(',')[0].trim()
}

function isSecureRequest(request) {
  const forwardedProto = getForwardedValue(request.headers['x-forwarded-proto'])
  if (forwardedProto) {
    return forwardedProto === 'https'
  }

  return Boolean(request.socket?.encrypted)
}

function getRequestHost(request) {
  return getForwardedValue(request.headers['x-forwarded-host']) || getForwardedValue(request.headers.host)
}

function buildResponseHeaders(request, extraHeaders = {}) {
  const headers = {
    'X-Content-Type-Options': 'nosniff',
    'X-Frame-Options': 'SAMEORIGIN',
    'Referrer-Policy': 'strict-origin-when-cross-origin',
    'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
    ...extraHeaders,
  }

  if (isSecureRequest(request)) {
    headers['Strict-Transport-Security'] = 'max-age=31536000; includeSubDomains'
  }

  return headers
}

function shouldRedirectToHttps(request) {
  if (!FORCE_HTTPS_REDIRECT) {
    return false
  }

  return !isSecureRequest(request)
}

function generateRuntimeConfig() {
  const configPath = path.join(DIST_DIR, 'config.js')
  const apiUrl = (process.env.VITE_API_URL || '/api').replace(/'/g, "\\'")

  writeFileSync(
    configPath,
    `window.ENV = {\n  VITE_API_URL: '${apiUrl}'\n};\n`,
    'utf8'
  )

  console.log(`Generated runtime config at ${configPath}`)
}

function sendFile(filePath, request, response) {
  const extension = path.extname(filePath).toLowerCase()
  const contentType = mimeTypes[extension] || 'application/octet-stream'
  const extraHeaders = {
    'Content-Type': contentType,
  }

  if (path.basename(filePath) === 'config.js') {
    extraHeaders['Cache-Control'] = 'no-store, no-cache, must-revalidate'
  }

  readFile(filePath, (error, content) => {
    if (error) {
      response.writeHead(500, buildResponseHeaders(request))
      response.end('Internal Server Error')
      return
    }

    response.writeHead(200, buildResponseHeaders(request, extraHeaders))
    response.end(content)
  })
}

async function proxyRequest(request, response) {
  const targetUrl = new URL(request.url, BACKEND_ORIGIN)
  const headers = new Headers()

  Object.entries(request.headers).forEach(([key, value]) => {
    if (!value || key.toLowerCase() === 'host') {
      return
    }

    if (Array.isArray(value)) {
      headers.set(key, value.join(', '))
      return
    }

    headers.set(key, value)
  })

  const host = getRequestHost(request)
  if (host) {
    headers.set('x-forwarded-host', host)
  }
  headers.set('x-forwarded-proto', isSecureRequest(request) ? 'https' : 'http')

  if (request.socket?.remoteAddress) {
    headers.set('x-forwarded-for', request.socket.remoteAddress)
  }

  const body =
    request.method === 'GET' || request.method === 'HEAD'
      ? undefined
      : await new Promise((resolve, reject) => {
          const chunks = []
          request.on('data', (chunk) => chunks.push(chunk))
          request.on('end', () => resolve(Buffer.concat(chunks)))
          request.on('error', reject)
        })

  const backendResponse = await fetch(targetUrl, {
    method: request.method,
    headers,
    body,
    redirect: 'manual',
  })

  const upstreamHeaders = Object.fromEntries(backendResponse.headers.entries())
  delete upstreamHeaders.connection
  delete upstreamHeaders['transfer-encoding']

  response.writeHead(backendResponse.status, buildResponseHeaders(request, upstreamHeaders))
  const arrayBuffer = await backendResponse.arrayBuffer()
  response.end(Buffer.from(arrayBuffer))
}

const server = createServer(async (request, response) => {
  try {
    if (!request.url) {
      response.writeHead(400, buildResponseHeaders(request))
      response.end('Bad Request')
      return
    }

    if (shouldRedirectToHttps(request)) {
      const host = getRequestHost(request)
      if (host) {
        response.writeHead(
          308,
          buildResponseHeaders(request, {
            Location: `https://${host}${request.url}`,
          })
        )
        response.end()
        return
      }
    }

    if (request.url.startsWith('/api') || request.url.startsWith('/uploads')) {
      await proxyRequest(request, response)
      return
    }

    const requestedPath = request.url === '/' ? '/index.html' : request.url
    const safePath = path.normalize(requestedPath).replace(/^(\.\.[/\\])+/, '')
    const filePath = path.join(DIST_DIR, safePath)

    if (existsSync(filePath) && statSync(filePath).isFile()) {
      sendFile(filePath, request, response)
      return
    }

    const nestedIndexPath = path.join(filePath, 'index.html')
    if (existsSync(nestedIndexPath)) {
      sendFile(nestedIndexPath, request, response)
      return
    }

    sendFile(path.join(DIST_DIR, 'index.html'), request, response)
  } catch (error) {
    response.writeHead(500, buildResponseHeaders(request))
    response.end(`Internal Server Error: ${error.message}`)
  }
})

generateRuntimeConfig()

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Frontend server is listening on port ${PORT}`)
  console.log(`HTTPS redirect is ${FORCE_HTTPS_REDIRECT ? 'enabled' : 'disabled'}`)
})
