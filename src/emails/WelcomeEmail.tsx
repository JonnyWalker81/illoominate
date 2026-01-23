import * as React from 'react';
import {
  Body,
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

interface WelcomeEmailProps {
  name?: string;
  position: number;
  referralCode: string;
}

export default function WelcomeEmail({
  name,
  position,
  referralCode,
}: WelcomeEmailProps) {
  return (
    <Html>
      <Head />
      <Preview>{`You're #${position} on the Illoominate waitlist!`}</Preview>
      <Body style={main}>
        <Container style={container}>
          <Heading style={heading}>
            Welcome{name ? `, ${name}` : ''}!
          </Heading>

          <Text style={paragraph}>
            You're <strong style={{ color: '#818cf8' }}>#{position}</strong> on the Illoominate waitlist.
            We're building native-first user feedback for indie developers
            and startups.
          </Text>

          <Hr style={hr} />

          <Section style={codeSection}>
            <Text style={codeLabel}>
              Move up the list! Share your referral code:
            </Text>
            <Text style={codeValue}>{referralCode}</Text>
            <Text style={codeHint}>
              Each friend who joins moves you up the list
            </Text>
          </Section>

          <Hr style={hr} />

          <Text style={footer}>
            Questions? Reply to this email.
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

const hr = {
  borderColor: '#333333',
  margin: '24px 0',
};

const codeSection = {
  textAlign: 'center' as const,
};

const codeLabel = {
  color: '#a1a1aa',
  fontSize: '14px',
  marginBottom: '8px',
};

const codeValue = {
  fontSize: '32px',
  fontWeight: 'bold' as const,
  fontFamily: 'monospace',
  color: '#818cf8',
  letterSpacing: '0.1em',
  margin: '8px 0',
};

const codeHint = {
  color: '#71717a',
  fontSize: '12px',
};

const footer = {
  color: '#71717a',
  fontSize: '12px',
  textAlign: 'center' as const,
};

const link = {
  color: '#818cf8',
};
