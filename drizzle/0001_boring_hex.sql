ALTER TABLE `quiz_responses` ADD `role` text;--> statement-breakpoint
ALTER TABLE `quiz_responses` ADD `disappointment_level` text;--> statement-breakpoint
ALTER TABLE `quiz_responses` ADD `quiz_completed` integer DEFAULT 0;