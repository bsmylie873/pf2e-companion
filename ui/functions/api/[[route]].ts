const BACKEND_ORIGIN = 'https://pf2e-companion.fly.dev'

export const onRequest: PagesFunction = async (context) => {
  const url = new URL(context.request.url)
  const path = url.pathname.replace(/^\/api/, '') || '/'
  const target = `${BACKEND_ORIGIN}${path}${url.search}`

  const headers = new Headers(context.request.headers)
  headers.delete('host')

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
