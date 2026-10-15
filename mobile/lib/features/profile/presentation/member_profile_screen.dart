import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/widgets/empty_state_widget.dart';
import '../../../core/widgets/error_state_widget.dart';
import '../../../core/widgets/skeleton_loader.dart';
import '../../groups/presentation/widgets/member_avatar.dart';
import '../domain/profile.dart';
import 'profile_notifier.dart';
import 'widgets/achievement_badge.dart';
import 'widgets/stat_card.dart';
import '../../../core/theme/app_tokens.dart';

class MemberProfileScreen extends ConsumerStatefulWidget {
  final String groupId;
  final String userId;

  const MemberProfileScreen({
    super.key,
    required this.groupId,
    required this.userId,
  });

  @override
  ConsumerState<MemberProfileScreen> createState() =>
      _MemberProfileScreenState();
}

class _MemberProfileScreenState extends ConsumerState<MemberProfileScreen> {
  late final _args = (groupId: widget.groupId, userId: widget.userId);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(profileProvider(_args).notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(profileProvider(_args));

    if (state.loading && state.profile == null) {
      return Scaffold(
        appBar: AppBar(),
        body: ListView(
          padding: const EdgeInsets.all(16),
          children: const [
            Row(
              children: [
                SkeletonLoader(width: 64, height: 64, borderRadius: 32),
                SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SkeletonLoader(width: 140, height: 24, borderRadius: 6),
                      SizedBox(height: 8),
                      SkeletonLoader(width: 200, height: 14, borderRadius: 6),
                    ],
                  ),
                ),
              ],
            ),
            SizedBox(height: 16),
            SkeletonLoader(height: 60, borderRadius: 12),
            SizedBox(height: 20),
            SkeletonLoader(height: 200, borderRadius: 12),
          ],
        ),
      );
    }

    if (state.error != null && state.profile == null) {
      return Scaffold(
        appBar: AppBar(),
        body: ErrorStateWidget(
          message: state.error,
          onRetry: () =>
              ref.read(profileProvider(_args).notifier).load(),
        ),
      );
    }

    final profile = state.profile;
    if (profile == null) return const SizedBox.shrink();

    return Scaffold(
      appBar: AppBar(
        title: Text(profile.user.username),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.read(profileProvider(_args).notifier).load(),
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _buildHeader(profile),
            const SizedBox(height: 16),
            _buildLegend(profile.legend),
            const SizedBox(height: 20),
            _buildStats(profile.stats),
            const SizedBox(height: 20),
            _buildAchievements(profile.achievements),
            if (profile.seasonHistory.isNotEmpty) ...[
              const SizedBox(height: 20),
              _buildSeasonHistory(profile.seasonHistory),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(MemberProfile profile) {
    return Row(
      children: [
        MemberAvatar(
          avatarEmoji: profile.user.avatarEmoji,
          avatarUrl: profile.user.avatarUrl,
          size: 64,
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                profile.user.username,
                style: context.ts.heading1.copyWith(fontSize: 24),
              ),
              if (profile.stats.topAttributeAllTime != null)
                Text(
                  profile.stats.topAttributeAllTime!.questionText,
                  style: context.ts.caption.copyWith(
                    color: context.t.color.accent,
                    fontWeight: FontWeight.w500,
                  ),
                ),
            ],
          ),
        ),
      ],
    ).animate().fadeIn(duration: context.motion(MotionClass.emphasis)).slideY(begin: -0.1, duration: context.motion(MotionClass.emphasis));
  }

  Widget _buildLegend(String legend) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: context.t.color.accentFill.withValues(alpha: 0.18),
        borderRadius: BorderRadius.circular(AppTokens.radius.md),
      ),
      child: Text(
        legend,
        style: context.ts.body.copyWith(
          fontStyle: FontStyle.italic,
          color: context.t.color.textPrimary,
        ),
      ),
    ).animate().fadeIn(duration: context.motion(MotionClass.emphasis), delay: AppTokens.motion.stagger(2));
  }

  Widget _buildStats(UserStats stats) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Статистика', style: context.ts.heading2.copyWith(fontSize: 18)),
        const SizedBox(height: 12),
        GridView.count(
          crossAxisCount: 2,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          mainAxisSpacing: 10,
          crossAxisSpacing: 10,
          childAspectRatio: 1.4,
          children: [
            StatCard(
              label: 'Сезонов сыграно',
              value: stats.seasonsPlayed.toString(),
              icon: Icons.calendar_today,
            ),
            StatCard(
              label: 'Стрик голосований',
              value: stats.votingStreak.toString(),
              icon: Icons.local_fire_department,
            ),
            // Shown only when the server sent it, which is only on one's own profile. Rendering a
            // fallback here would turn "withheld" into "0% — never guessed right", which is a
            // different and false claim about another member.
            if (stats.guessAccuracy != null)
              StatCard(
                label: 'Точность угадывания',
                value: '${stats.guessAccuracy}%',
                icon: Icons.track_changes,
              ),
            StatCard(
              label: 'Голосов получено',
              value: stats.totalVotesReceived.toString(),
              icon: Icons.how_to_vote,
            ),
            StatCard(
              label: stats.topAttributeAllTime != null
                  ? stats.topAttributeAllTime!.questionText
                  : 'Лучший атрибут',
              value: stats.topAttributeAllTime != null
                  ? '${stats.topAttributeAllTime!.percentage}%'
                  : '-',
              icon: Icons.star,
              animateNumber: stats.topAttributeAllTime != null,
            ),
            StatCard(
              label: 'Макс. стрик',
              value: stats.maxVotingStreak.toString(),
              icon: Icons.emoji_events,
            ),
          ],
        ),
      ],
    ).animate().fadeIn(duration: context.motion(MotionClass.emphasis), delay: AppTokens.motion.stagger(3));
  }

  Widget _buildAchievements(List<ProfileAchievement> achievements) {
    if (achievements.isEmpty) {
      return const EmptyStateWidget(
        emoji: '\u{1F3C6}',
        title: 'Пока нет ачивок',
        subtitle: 'Голосуй и узнай что о тебе думают',
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Ачивки', style: context.ts.heading2.copyWith(fontSize: 18)),
        const SizedBox(height: 12),
        SizedBox(
          height: 120,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            itemCount: achievements.length,
            separatorBuilder: (_, _) => const SizedBox(width: 10),
            itemBuilder: (context, index) {
              final a = achievements[index];
              return AchievementBadge(
                type: a.type,
                earnedAt: a.earnedAt,
                unlocked: true,
              );
            },
          ),
        ),
      ],
    ).animate().fadeIn(duration: context.motion(MotionClass.emphasis), delay: AppTokens.motion.stagger(5));
  }

  Widget _buildSeasonHistory(List<SeasonCardDto> history) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('История сезонов',
            style: context.ts.heading2.copyWith(fontSize: 18)),
        const SizedBox(height: 12),
        ...List.generate(history.length, (index) {
          final card = history[index];
          return Container(
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
        color: context.t.elevation.level1.surface,
        borderRadius: AppTokens.radius.card,
        border: Border.all(
          color: context.t.elevation.level1.outline,
          width: AppTokens.border.hairline,
        ),
        boxShadow: context.t.elevation.level1.shadows,
      ),
            child: Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: context.t.color.accentFill.withValues(alpha: 0.18),
                    borderRadius: BorderRadius.circular(AppTokens.radius.sm),
                  ),
                  alignment: Alignment.center,
                  child: Text(
                    '#${card.seasonNumber}',
                    style: context.ts.caption.copyWith(
                      fontWeight: FontWeight.bold,
                      color: context.t.color.accent,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        card.topAttribute,
                        style: context.ts.body.copyWith(
                          fontWeight: FontWeight.w500,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        '${card.percentage}%',
                        style: context.ts.caption.copyWith(
                          color: context.t.color.accent,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          )
              .animate()
              .fadeIn(
                  duration: context.motion(MotionClass.surface),
                  delay: AppTokens.motion.stagger(5 + index))
              .slideX(begin: 0.1, duration: context.motion(MotionClass.surface));
        }),
      ],
    );
  }
}
