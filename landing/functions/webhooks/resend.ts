import type { Env, WaitlistEntry } from '../types.d.ts';

interface ResendWebhookEvent {
  type: 'email.sent' | 'email.delivered' | 'email.opened' | 'email.clicked' | 'email.bounced' | 'email.complained' | 'email.delivery_delayed';
  created_at: string;
  data: {
    email_id: string;
    to: string[];
    from: string;
    subject: string;
    click?: {
      link: string;
      timestamp: string;
    };
  };
}

export async function onRequestPost(context: {
  request: Request;
  env: Env;
}): Promise<Response> {
  const { request, env } = context;

  try {
    // Verify webhook signature using Svix
    const svixId = request.headers.get('svix-id');
    const svixTimestamp = request.headers.get('svix-timestamp');
    const svixSignature = request.headers.get('svix-signature');

    if (!svixId || !svixTimestamp || !svixSignature) {
      return new Response('Missing webhook headers', { status: 400 });
    }

    const body = await request.text();

    // Verify signature (simplified - in production, use @svix/webhook)
    const isValid = await verifyWebhookSignature(
      body,
      svixId,
      svixTimestamp,
      svixSignature,
      env.RESEND_WEBHOOK_SECRET
    );

    if (!isValid) {
      return new Response('Invalid signature', { status: 401 });
    }

    const event = JSON.parse(body) as ResendWebhookEvent;
    const email = event.data.to[0];

    // Find waitlist entry
    const entry = await env.DB.prepare(
      'SELECT id FROM waitlist WHERE email = ?'
    )
      .bind(email.toLowerCase())
      .first<Pick<WaitlistEntry, 'id'>>();

    if (!entry) {
      // Email not in our system - ignore
      return new Response('OK', { status: 200 });
    }

    // Map event type to our schema
    const eventTypeMap: Record<string, string> = {
      'email.sent': 'sent',
      'email.delivered': 'delivered',
      'email.opened': 'opened',
      'email.clicked': 'clicked',
      'email.bounced': 'bounced',
      'email.complained': 'complained',
      'email.delivery_delayed': 'delivery_delayed',
    };

    const eventType = eventTypeMap[event.type];
    if (!eventType) {
      return new Response('Unknown event type', { status: 200 });
    }

    // Store email event
    await env.DB.prepare(
      `INSERT INTO email_events (
        waitlist_id, event_type, resend_email_id, link_url,
        event_timestamp, raw_webhook_data
      ) VALUES (?, ?, ?, ?, ?, ?)`
    )
      .bind(
        entry.id,
        eventType,
        event.data.email_id,
        event.data.click?.link || null,
        event.created_at,
        body
      )
      .run();

    // Update engagement metrics based on event type
    switch (eventType) {
      case 'sent':
        await env.DB.prepare(
          'UPDATE waitlist SET total_emails_sent = total_emails_sent + 1 WHERE id = ?'
        )
          .bind(entry.id)
          .run();
        break;

      case 'opened':
        await env.DB.prepare(
          `UPDATE waitlist SET
            total_emails_opened = total_emails_opened + 1,
            last_email_opened_at = datetime('now'),
            engagement_score = engagement_score + 1
          WHERE id = ?`
        )
          .bind(entry.id)
          .run();
        break;

      case 'clicked':
        await env.DB.prepare(
          `UPDATE waitlist SET
            total_emails_clicked = total_emails_clicked + 1,
            last_email_clicked_at = datetime('now'),
            engagement_score = engagement_score + 2
          WHERE id = ?`
        )
          .bind(entry.id)
          .run();
        break;

      case 'bounced':
        await env.DB.prepare(
          "UPDATE waitlist SET email_status = 'bounced' WHERE id = ?"
        )
          .bind(entry.id)
          .run();
        break;

      case 'complained':
        await env.DB.prepare(
          `UPDATE waitlist SET
            email_status = 'unsubscribed',
            unsubscribed_at = datetime('now'),
            unsubscribe_reason = 'spam_complaint'
          WHERE id = ?`
        )
          .bind(entry.id)
          .run();
        break;
    }

    return new Response('OK', { status: 200 });
  } catch (error) {
    console.error('Webhook error:', error);
    return new Response('Internal error', { status: 500 });
  }
}

async function verifyWebhookSignature(
  payload: string,
  svixId: string,
  svixTimestamp: string,
  svixSignature: string,
  secret: string
): Promise<boolean> {
  try {
    // Check timestamp is within 5 minutes
    const timestamp = parseInt(svixTimestamp, 10);
    const now = Math.floor(Date.now() / 1000);
    if (Math.abs(now - timestamp) > 300) {
      return false;
    }

    // Extract the secret key (remove whsec_ prefix if present)
    const secretKey = secret.startsWith('whsec_') ? secret.slice(6) : secret;
    const secretBytes = base64ToUint8Array(secretKey);

    // Create the signature base string
    const signedContent = `${svixId}.${svixTimestamp}.${payload}`;
    const encoder = new TextEncoder();
    const data = encoder.encode(signedContent);

    // Import key and sign
    const key = await crypto.subtle.importKey(
      'raw',
      secretBytes,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );

    const signature = await crypto.subtle.sign('HMAC', key, data);
    const expectedSignature = uint8ArrayToBase64(new Uint8Array(signature));

    // Parse the signature header (format: v1,signature1 v1,signature2 ...)
    const signatures = svixSignature.split(' ').map((s) => s.split(',')[1]);

    return signatures.some((sig) => sig === expectedSignature);
  } catch (error) {
    console.error('Signature verification error:', error);
    return false;
  }
}

function base64ToUint8Array(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}

function uint8ArrayToBase64(bytes: Uint8Array): string {
  let binary = '';
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}
