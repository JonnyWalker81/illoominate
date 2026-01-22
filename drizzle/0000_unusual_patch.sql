CREATE TABLE `quiz_responses` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`waitlist_id` integer,
	`platform` text,
	`team_size` text,
	`pain_points` text,
	`created_at` text NOT NULL,
	FOREIGN KEY (`waitlist_id`) REFERENCES `waitlist`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `waitlist` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`email` text NOT NULL,
	`name` text,
	`source` text,
	`referral_code` text NOT NULL,
	`referred_by` text,
	`referral_count` integer DEFAULT 0,
	`created_at` text NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `waitlist_email_unique` ON `waitlist` (`email`);--> statement-breakpoint
CREATE UNIQUE INDEX `waitlist_referral_code_unique` ON `waitlist` (`referral_code`);