import * as React from 'react';
import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Html,
  Preview,
  Section,
  Text,
  Hr,
  Link,
} from '@react-email/components';

interface VerificationEmailProps {
  name?: string;
  verificationUrl: string;
}

export default function VerificationEmail({
  name,
  verificationUrl,
}: VerificationEmailProps) {
  return (
    <Html>
      <Head />
      <Preview>Verify your email to join the Illoominate waitlist</Preview>
      <Body style={main}>
        <Container style={container}>
          <Heading style={heading}>
            Verify your email{name ? `, ${name}` : ''}
          </Heading>

          <Text style={paragraph}>
            Thanks for your interest in Illoominate! Click the button below to
            verify your email and secure your spot on the waitlist.
          </Text>

          <Section style={buttonSection}>
            <Button style={button} href={verificationUrl}>
              Verify Email
            </Button>
          </Section>

          <Text style={linkFallback}>
            Or copy and paste this link:
            <br />
            <Link href={verificationUrl} style={link}>
              {verificationUrl}
            </Link>
          </Text>

          <Hr style={hr} />

          <Text style={footer}>
            This link expires in 24 hours.
            <br />
            If you didn't sign up for Illoominate, you can ignore this email.
            <br />
            <Link href="https://illoominate.app" style={link}>
              illoominate.app
            </Link>
          </Text>
        </Container>
      </Body>
    </Html>
  );
}

// Styles
const main = {
  backgroundColor: '#0a0a0a',
  color: '#ffffff',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};

const container = {
  padding: '40px 20px',
  maxWidth: '480px',
  margin: '0 auto',
};

const heading = {
  fontSize: '24px',
  fontWeight: '600' as const,
  marginBottom: '16px',
  color: '#ffffff',
};

const paragraph = {
  color: '#a1a1aa',
  lineHeight: '1.6',
  fontSize: '16px',
};

const buttonSection = {
  textAlign: 'center' as const,
  margin: '32px 0',
};

const button = {
  backgroundColor: '#818cf8',
  color: '#ffffff',
  padding: '14px 28px',
  borderRadius: '8px',
  fontSize: '16px',
  fontWeight: '600' as const,
  textDecoration: 'none',
};

const linkFallback = {
  color: '#71717a',
  fontSize: '12px',
  textAlign: 'center' as const,
  wordBreak: 'break-all' as const,
};

const hr = {
  borderColor: '#333333',
  margin: '24px 0',
};

const footer = {
  color: '#71717a',
  fontSize: '12px',
  textAlign: 'center' as const,
};

const link = {
  color: '#818cf8',
};
