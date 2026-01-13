import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Waitlist from '../Waitlist';

describe('Waitlist Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Initial Render', () => {
    it('should render the waitlist section', () => {
      render(<Waitlist />);

      expect(screen.getByText('Be First in Line')).toBeInTheDocument();
    });

    it('should render email input field', () => {
      render(<Waitlist />);

      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    });

    it('should render name input field', () => {
      render(<Waitlist />);

      expect(screen.getByLabelText(/^name/i)).toBeInTheDocument();
    });

    it('should render referral source dropdown', () => {
      render(<Waitlist />);

      expect(screen.getByLabelText(/how did you hear about us/i)).toBeInTheDocument();
    });

    it('should render submit button', () => {
      render(<Waitlist />);

      expect(screen.getByRole('button', { name: /join the waitlist/i })).toBeInTheDocument();
    });

    it('should show email field as required', () => {
      render(<Waitlist />);

      const emailInput = screen.getByLabelText(/email address/i);
      expect(emailInput).toBeRequired();
    });
  });

  describe('Form Validation', () => {
    it('should show validation error for invalid email on blur', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      const emailInput = screen.getByLabelText(/email address/i);
      await user.type(emailInput, 'invalid-email');
      await user.tab(); // Blur the input

      // HTML5 validation would handle this, but we can check the input validity
      expect((emailInput as HTMLInputElement).validity.valid).toBe(false);
    });

    it('should accept valid email format', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      const emailInput = screen.getByLabelText(/email address/i);
      await user.type(emailInput, 'test@example.com');

      expect((emailInput as HTMLInputElement).validity.valid).toBe(true);
    });
  });

  describe('Form Submission', () => {
    it('should show loading state during submission', async () => {
      // Mock fetch to delay response
      vi.mocked(globalThis.fetch).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(new Response(JSON.stringify({ success: true, invite_code: 'ABC12345' }))), 1000))
      );

      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByText('Joining...')).toBeInTheDocument();
    });

    it('should call API on form submission', async () => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({ success: true, invite_code: 'ABC12345', message: 'Success' }))
      );

      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      await user.type(screen.getByLabelText(/^name/i), 'Test User');
      await user.selectOptions(screen.getByLabelText(/how did you hear about us/i), 'reddit');

      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        expect(globalThis.fetch).toHaveBeenCalledWith('/api/waitlist/join', expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
        }));
      });
    });

    it('should include all form data in API request', async () => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({ success: true, invite_code: 'ABC12345', message: 'Success' }))
      );

      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      await user.type(screen.getByLabelText(/^name/i), 'Test User');
      await user.selectOptions(screen.getByLabelText(/how did you hear about us/i), 'friend');

      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        const [, options] = vi.mocked(globalThis.fetch).mock.calls[0];
        const body = JSON.parse(options?.body as string);

        expect(body.email).toBe('test@example.com');
        expect(body.name).toBe('Test User');
        expect(body.referral_source).toBe('friend');
      });
    });
  });

  describe('Success State', () => {
    beforeEach(() => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({
          success: true,
          invite_code: 'TESTCODE',
          message: 'Check your email to verify',
        }))
      );
    });

    it('should show success message after submission', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByText("You're on the list!")).toBeInTheDocument();
    });

    it('should display invite code', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        // The referral link contains the invite code
        const input = screen.getByDisplayValue(/TESTCODE/i);
        expect(input).toBeInTheDocument();
      });
    });

    it('should show referral link with invite code', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        expect(screen.getByText('Your referral link:')).toBeInTheDocument();
      });
    });

    it('should show Copy button', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByRole('button', { name: 'Copy' })).toBeInTheDocument();
    });

    it('should show Share on Twitter button', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByText('Share on Twitter')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should display error message on API failure', async () => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({ error: 'This email is already on the waitlist.' }), { status: 409 })
      );

      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'duplicate@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByText('This email is already on the waitlist.')).toBeInTheDocument();
    });

    it('should show generic error on network failure', async () => {
      vi.mocked(globalThis.fetch).mockRejectedValue(new Error('Network error'));

      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByText('Network error')).toBeInTheDocument();
    });
  });

  describe('Referral Code', () => {
    it('should accept referral code from props', () => {
      render(<Waitlist referralCode="REFCODE1" />);

      expect(screen.getByDisplayValue('REFCODE1')).toBeInTheDocument();
    });

    it('should include referral code in API request', async () => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({ success: true, invite_code: 'ABC12345', message: 'Success' }))
      );

      render(<Waitlist referralCode="REFCODE1" />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        const [, options] = vi.mocked(globalThis.fetch).mock.calls[0];
        const body = JSON.parse(options?.body as string);

        expect(body.referral_code).toBe('REFCODE1');
      });
    });
  });

  describe('Copy Functionality', () => {
    const mockWriteText = vi.fn().mockResolvedValue(undefined);

    beforeEach(() => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({
          success: true,
          invite_code: 'TESTCODE',
          message: 'Success',
        }))
      );

      // Mock clipboard using defineProperty
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: mockWriteText,
        },
        writable: true,
        configurable: true,
      });
    });

    it('should have Copy button in success state', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      expect(await screen.findByRole('button', { name: 'Copy' })).toBeInTheDocument();
    });

    it('should show Copied! after clicking copy', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await screen.findByRole('button', { name: 'Copy' });
      fireEvent.click(screen.getByRole('button', { name: 'Copy' }));

      expect(await screen.findByText('Copied!')).toBeInTheDocument();
    });
  });

  describe('Referral Source Options', () => {
    it('should have all referral source options', () => {
      render(<Waitlist />);

      const select = screen.getByLabelText(/how did you hear about us/i);

      expect(select).toContainHTML('Reddit');
      expect(select).toContainHTML('Search Engine');
      expect(select).toContainHTML('Social Media');
      expect(select).toContainHTML('Friend or Colleague');
      expect(select).toContainHTML('Blog or Article');
      expect(select).toContainHTML('Other');
    });
  });

  describe('Quiz in Success State', () => {
    beforeEach(() => {
      vi.mocked(globalThis.fetch).mockResolvedValue(
        new Response(JSON.stringify({
          success: true,
          invite_code: 'TESTCODE',
          message: 'Success',
        }))
      );
    });

    it('should show quiz section after signup', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        expect(screen.getByText('While you wait...')).toBeInTheDocument();
      });
    });

    it('should show skip quiz link', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        expect(screen.getByText('Skip quiz →')).toBeInTheDocument();
      });
    });

    it('should hide quiz when skip is clicked', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        expect(screen.getByText('Skip quiz →')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Skip quiz →'));

      await waitFor(() => {
        expect(screen.queryByText('While you wait...')).not.toBeInTheDocument();
        expect(screen.queryByText('Skip quiz →')).not.toBeInTheDocument();
      });
    });

    it('should show quiz questions in success state', async () => {
      render(<Waitlist />);
      const user = userEvent.setup();

      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      fireEvent.click(screen.getByRole('button', { name: /join the waitlist/i }));

      await waitFor(() => {
        // Quiz should show first question
        expect(screen.getByText('How do you currently collect user feedback?')).toBeInTheDocument();
      });
    });
  });
});
