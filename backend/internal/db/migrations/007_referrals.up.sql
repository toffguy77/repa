-- 007_referrals.up.sql
-- Who brought a member in, so a referral reward reaches the person who actually invited rather
-- than the group's admin. Separate from join_source: the channel and the referrer are different
-- facts, and a channel can exist without a referrer.

ALTER TABLE group_members
  ADD COLUMN invited_by TEXT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_group_members_invited_by ON group_members(invited_by);
