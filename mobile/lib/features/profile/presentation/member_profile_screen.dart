import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../groups/presentation/widgets/member_avatar.dart';
import '../domain/profile.dart';
import 'profile_notifier.dart';
import 'widgets/achievement_badge.dart';
import 'widgets/stat_card.dart';

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
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (state.error != null && state.profile == null) {
      return Scaffold(
        appBar: AppBar(),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(state.error!, style: AppTextStyles.body),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () =>
                    ref.read(profileProvider(_args).notifier).load(),
                child: const Text('Повторить'),
              ),
            ],
          ),
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
            if (profile.achievements.isNotEmpty) ...[
              _buildAchievements(profile.achievements),
              const SizedBox(height: 20),
            ],
            if (profile.seasonHistory.isNotEmpty)
              _buildSeasonHistory(profile.seasonHistory),
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
                style: AppTextStyles.headline1.copyWith(fontSize: 24),
              ),
              if (profile.stats.topAttributeAllTime != null)
                Text(
                  profile.stats.topAttributeAllTime!.questionText,
                  style: AppTextStyles.caption.copyWith(
                    color: AppColors.primary,
                    fontWeight: FontWeight.w500,
                  ),
                ),
            ],
          ),
        ),
      ],
    ).animate().fadeIn(duration: 400.ms).slideY(begin: -0.1, duration: 400.ms);
  }

  Widget _buildLegend(String legend) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.primaryLight,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        legend,
        style: AppTextStyles.body.copyWith(
          fontStyle: FontStyle.italic,
          color: AppColors.textPrimary,
        ),
      ),
    ).animate().fadeIn(duration: 400.ms, delay: 100.ms);
  }

  Widget _buildStats(UserStats stats) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Статистика', style: AppTextStyles.headline2.copyWith(fontSize: 18)),
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
    ).animate().fadeIn(duration: 400.ms, delay: 200.ms);
  }

  Widget _buildAchievements(List<ProfileAchievement> achievements) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Ачивки', style: AppTextStyles.headline2.copyWith(fontSize: 18)),
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
    ).animate().fadeIn(duration: 400.ms, delay: 300.ms);
  }

  Widget _buildSeasonHistory(List<SeasonCardDto> history) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('История сезонов',
            style: AppTextStyles.headline2.copyWith(fontSize: 18)),
        const SizedBox(height: 12),
        ...List.generate(history.length, (index) {
          final card = history[index];
          return Container(
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: AppColors.primaryLight,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  alignment: Alignment.center,
                  child: Text(
                    '#${card.seasonNumber}',
                    style: AppTextStyles.caption.copyWith(
                      fontWeight: FontWeight.bold,
                      color: AppColors.primary,
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
                        style: AppTextStyles.body.copyWith(
                          fontWeight: FontWeight.w500,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        '${card.percentage}%',
                        style: AppTextStyles.caption.copyWith(
                          color: AppColors.primary,
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
                  duration: 300.ms,
                  delay: Duration(milliseconds: 400 + index * 80))
              .slideX(begin: 0.1, duration: 300.ms);
        }),
      ],
    );
  }
}
