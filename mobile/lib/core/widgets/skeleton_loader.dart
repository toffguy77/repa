import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../theme/app_colors.dart';

class SkeletonLoader extends StatelessWidget {
  final double width;
  final double height;
  final double borderRadius;

  const SkeletonLoader({
    super.key,
    this.width = double.infinity,
    required this.height,
    this.borderRadius = 12,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: isDark ? Colors.grey.shade800 : AppColors.surface,
        borderRadius: BorderRadius.circular(borderRadius),
      ),
    )
        .animate(onPlay: (c) => c.repeat())
        .shimmer(
          duration: 1200.ms,
          color: (isDark ? Colors.grey.shade600 : Colors.white)
              .withValues(alpha: 0.5),
        );
  }
}

class GroupCardSkeleton extends StatelessWidget {
  const GroupCardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              SkeletonLoader(width: 44, height: 44, borderRadius: 22),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SkeletonLoader(width: 140, height: 18, borderRadius: 6),
                    SizedBox(height: 6),
                    SkeletonLoader(width: 90, height: 14, borderRadius: 6),
                  ],
                ),
              ),
            ],
          ),
          SizedBox(height: 14),
          SkeletonLoader(height: 8, borderRadius: 4),
          SizedBox(height: 10),
          SkeletonLoader(width: 120, height: 14, borderRadius: 6),
        ],
      ),
    );
  }
}

class MemberAvatarSkeleton extends StatelessWidget {
  const MemberAvatarSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          SkeletonLoader(width: 44, height: 44, borderRadius: 22),
          SizedBox(width: 12),
          Expanded(
            child: SkeletonLoader(width: 120, height: 16, borderRadius: 6),
          ),
        ],
      ),
    );
  }
}

class MemberCardSkeleton extends StatelessWidget {
  const MemberCardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              SkeletonLoader(width: 44, height: 44, borderRadius: 22),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SkeletonLoader(width: 100, height: 16, borderRadius: 6),
                    SizedBox(height: 6),
                    SkeletonLoader(width: 140, height: 14, borderRadius: 6),
                  ],
                ),
              ),
            ],
          ),
          SizedBox(height: 16),
          SkeletonLoader(height: 12, borderRadius: 4),
          SizedBox(height: 10),
          SkeletonLoader(height: 12, borderRadius: 4),
          SizedBox(height: 10),
          SkeletonLoader(width: 180, height: 12, borderRadius: 4),
        ],
      ),
    );
  }
}

class VotingSessionSkeleton extends StatelessWidget {
  const VotingSessionSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        children: [
          // Progress bar
          SkeletonLoader(height: 6, borderRadius: 4),
          SizedBox(height: 24),
          // Question card
          SkeletonLoader(height: 100, borderRadius: 16),
          SizedBox(height: 24),
          // Participant grid (2x2)
          Row(
            children: [
              Expanded(child: SkeletonLoader(height: 120, borderRadius: 16)),
              SizedBox(width: 12),
              Expanded(child: SkeletonLoader(height: 120, borderRadius: 16)),
            ],
          ),
          SizedBox(height: 12),
          Row(
            children: [
              Expanded(child: SkeletonLoader(height: 120, borderRadius: 16)),
              SizedBox(width: 12),
              Expanded(child: SkeletonLoader(height: 120, borderRadius: 16)),
            ],
          ),
        ],
      ),
    );
  }
}
