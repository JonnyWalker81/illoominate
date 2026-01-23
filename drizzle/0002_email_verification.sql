ALTER TABLE `waitlist` ADD `verified` integer DEFAULT 0;--> statement-breakpoint
ALTER TABLE `waitlist` ADD `verification_token` text;--> statement-breakpoint
ALTER TABLE `waitlist` ADD `verified_at` text;
