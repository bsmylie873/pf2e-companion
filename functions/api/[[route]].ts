interface Env {
  BACKEND_ORIGIN: string
}

export const onRequest: PagesFunction<Env> = async (context) => {
  const url = new URL(context.request.url)
  const path = url.pathname.replace(/^\/api/, '') || '/'
  const target = `${context.env.BACKEND_ORIGIN}${path}${url.search}`

  const headers = new Headers(context.request.headers)
  headers.delete('host')

  if (context.request.headers.get('Upgrade') === 'websocket') {
    return fetch(target, { headers })
  }

  const res = await fetch(target, {
    method: context.request.method,
    headers,
    body: context.request.body,
    redirect: 'manual',
  })

  return new Response(res.body, {
    status: res.status,
    statusText: res.statusText,
    headers: res.headers,
  })
}
