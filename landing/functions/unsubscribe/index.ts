import type { Env, WaitlistEntry } from '../types.d.ts';

export async function onRequestGet(context: {
  request: Request;
  env: Env;
}): Promise<Response> {
  const { request, env } = context;

  try {
    const url = new URL(request.url);
    const token = url.searchParams.get('token');

    if (!token) {
      return htmlResponse(renderPage('error', 'Invalid unsubscribe link.'));
    }

    // Find entry by unsubscribe token
    const entry = await env.DB.prepare(
      'SELECT id, email, email_status FROM waitlist WHERE unsubscribe_token = ?'
    )
      .bind(token)
      .first<Pick<WaitlistEntry, 'id' | 'email' | 'email_status'>>();

    if (!entry) {
      return htmlResponse(renderPage('error', 'Invalid unsubscribe link.'));
    }

    if (entry.email_status === 'unsubscribed') {
      return htmlResponse(renderPage('already', 'You have already been unsubscribed.'));
    }

    // Mark as unsubscribed
    await env.DB.prepare(
      `UPDATE waitlist SET
        email_status = 'unsubscribed',
        unsubscribed_at = datetime('now'),
        marketing_consent = false
      WHERE id = ?`
    )
      .bind(entry.id)
      .run();

    // Remove from Resend audience
    await removeFromResend(env, entry.email);

    return htmlResponse(renderPage('success', 'You have been unsubscribed.'));
  } catch (error) {
    console.error('Unsubscribe error:', error);
    return htmlResponse(renderPage('error', 'Something went wrong. Please try again.'));
  }
}

async function removeFromResend(env: Env, email: string): Promise<void> {
  try {
    // First, get the contact ID
    const response = await fetch(
      `https://api.resend.com/audiences/${env.RESEND_AUDIENCE_ID}/contacts?email=${encodeURIComponent(email)}`,
      {
        headers: {
          Authorization: `Bearer ${env.RESEND_API_KEY}`,
        },
      }
    );

    if (response.ok) {
      const data = await response.json() as { data: Array<{ id: string }> };
      if (data.data && data.data.length > 0) {
        const contactId = data.data[0].id;

        // Delete the contact
        await fetch(
          `https://api.resend.com/audiences/${env.RESEND_AUDIENCE_ID}/contacts/${contactId}`,
          {
            method: 'DELETE',
            headers: {
              Authorization: `Bearer ${env.RESEND_API_KEY}`,
            },
          }
        );
      }
    }
  } catch (error) {
    console.error('Failed to remove from Resend:', error);
  }
}

function renderPage(status: 'success' | 'error' | 'already', message: string): string {
  const icon = status === 'success'
    ? '<svg class="icon success" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>'
    : '<svg class="icon error" viewBox="0 0 24 24"><path d="M6 18L18 6M6 6l12 12"/></svg>';

  const color = status === 'success' ? '#10B981' : '#EF4444';

  return `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Unsubscribe - Illoominate</title>
      <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
          background: linear-gradient(135deg, #f8fafc 0%, #eef2ff 100%);
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
        }
        .container {
          background: white;
          border-radius: 16px;
          padding: 48px;
          text-align: center;
          box-shadow: 0 10px 40px rgba(0,0,0,0.1);
          max-width: 400px;
          width: 100%;
        }
        .icon {
          width: 64px;
          height: 64px;
          margin-bottom: 24px;
          stroke: ${color};
          stroke-width: 2;
          fill: none;
          stroke-linecap: round;
          stroke-linejoin: round;
        }
        h1 {
          color: #1e293b;
          font-size: 24px;
          margin-bottom: 12px;
        }
        p {
          color: #64748b;
          line-height: 1.6;
        }
        .btn {
          display: inline-block;
          margin-top: 24px;
          padding: 12px 24px;
          background: #4F46E5;
          color: white;
          text-decoration: none;
          border-radius: 8px;
          font-weight: 500;
        }
        .btn:hover { background: #4338CA; }
      </style>
    </head>
    <body>
      <div class="container">
        ${icon}
        <h1>${status === 'success' ? 'Unsubscribed' : 'Oops'}</h1>
        <p>${message}</p>
        <a href="/" class="btn">Return Home</a>
      </div>
    </body>
    </html>
  `;
}

function htmlResponse(html: string, status = 200): Response {
  return new Response(html, {
    status,
    headers: {
      'Content-Type': 'text/html; charset=utf-8',
    },
  });
}
