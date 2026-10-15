import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../theme/app_tokens.dart';
import 'kit/kit.dart';

/// A shimmering placeholder block.
///
/// Radius defaults to the small step so a placeholder matches the shape of whatever it
/// stands in for; pass [borderRadius] for pills and avatars.
class SkeletonLoader extends StatelessWidget {
  final double width;
  final double height;
  final double borderRadius;

  const SkeletonLoader({
    super.key,
    this.width = double.infinity,
    required this.height,
    this.borderRadius = 6,
  });

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final placeholder = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: t.color.surface,
        borderRadius: BorderRadius.circular(borderRadius),
      ),
    );

    // A shimmer conveys nothing the static placeholder does not, so reduced motion drops
    // it entirely rather than animating it at zero duration.
    if (MediaQuery.disableAnimationsOf(context)) return placeholder;

    return placeholder.animate(onPlay: (c) => c.repeat()).shimmer(
          duration: AppTokens.motion.shimmerPeriod,
          color: t.color.textSecondary.withValues(alpha: 0.25),
        );
  }
}

class GroupCardSkeleton extends StatelessWidget {
  const GroupCardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.symmetric(
        horizontal: AppTokens.space.lg,
        vertical: AppTokens.space.xs,
      ),
      child: AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const SkeletonLoader(width: 44, height: 44, borderRadius: 22),
                SizedBox(width: AppTokens.space.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SkeletonLoader(width: 140, height: 18),
                      SizedBox(height: AppTokens.space.xs),
                      const SkeletonLoader(width: 90, height: 14),
                    ],
                  ),
                ),
              ],
            ),
            SizedBox(height: AppTokens.space.md),
            const SkeletonLoader(height: 6, borderRadius: 999),
            SizedBox(height: AppTokens.space.sm),
            const SkeletonLoader(width: 120, height: 14),
          ],
        ),
      ),
    );
  }
}

class MemberAvatarSkeleton extends StatelessWidget {
  const MemberAvatarSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: AppTokens.space.sm),
      child: Row(
        children: [
          const SkeletonLoader(width: 44, height: 44, borderRadius: 22),
          SizedBox(width: AppTokens.space.md),
          const Expanded(child: SkeletonLoader(width: 120, height: 16)),
        ],
      ),
    );
  }
}

class MemberCardSkeleton extends StatelessWidget {
  const MemberCardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: AppTokens.space.md),
      child: AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const SkeletonLoader(width: 44, height: 44, borderRadius: 22),
                SizedBox(width: AppTokens.space.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SkeletonLoader(width: 100, height: 16),
                      SizedBox(height: AppTokens.space.xs),
                      const SkeletonLoader(width: 140, height: 14),
                    ],
                  ),
                ),
              ],
            ),
            SizedBox(height: AppTokens.space.lg),
            const SkeletonLoader(height: 12),
            SizedBox(height: AppTokens.space.sm),
            const SkeletonLoader(height: 12),
            SizedBox(height: AppTokens.space.sm),
            const SkeletonLoader(width: 180, height: 12),
          ],
        ),
      ),
    );
  }
}

class VotingSessionSkeleton extends StatelessWidget {
  const VotingSessionSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    final cardRadius = AppTokens.radius.lg;
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: AppTokens.space.lg),
      child: Column(
        children: [
          const SkeletonLoader(height: 6, borderRadius: 999),
          SizedBox(height: AppTokens.space.xl),
          SkeletonLoader(height: 100, borderRadius: cardRadius),
          SizedBox(height: AppTokens.space.xl),
          Row(
            children: [
              Expanded(child: SkeletonLoader(height: 120, borderRadius: cardRadius)),
              SizedBox(width: AppTokens.space.md),
              Expanded(child: SkeletonLoader(height: 120, borderRadius: cardRadius)),
            ],
          ),
          SizedBox(height: AppTokens.space.md),
          Row(
            children: [
              Expanded(child: SkeletonLoader(height: 120, borderRadius: cardRadius)),
              SizedBox(width: AppTokens.space.md),
              Expanded(child: SkeletonLoader(height: 120, borderRadius: cardRadius)),
            ],
          ),
        ],
      ),
    );
  }
}
