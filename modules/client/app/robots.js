export default function robots() {
  return {
    rules: [
      {
        userAgent: '*',
        allow: '/',
      },
      {
        userAgent: ['GPTBot', 'ChatGPT-User', 'Google-Extended', 'CCBot', 'anthropic-ai', 'Claude-Web', 'Bytespider', 'cohere-ai'],
        disallow: '/',
      },
    ],
  }
}
