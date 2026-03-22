import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../domain/voting.dart';
import 'voting_notifier.dart';

class VotingCompleteScreen extends ConsumerStatefulWidget {
  final String groupId;
  final String seasonId;

  const VotingCompleteScreen({
    super.key,
    required this.groupId,
    required this.seasonId,
  });

  @override
  ConsumerState<VotingCompleteScreen> createState() =>
      _VotingCompleteScreenState();
}

class _VotingCompleteScreenState extends ConsumerState<VotingCompleteScreen> {
  @override
  void initState() {
    super.initState();
    HapticFeedback.heavyImpact();
  }

  @override
  Widget build(BuildContext context) {
    final progressAsync =
        ref.watch(groupVotingProgressProvider(widget.seasonId));

    return PopScope(
      canPop: false,
      child: Scaffold(
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              children: [
                const Spacer(),

                const Text(
                  '\u{1F389}',
                  style: TextStyle(fontSize: 80),
                )
                    .animate()
                    .scale(
                      begin: const Offset(0.3, 0.3),
                      end: const Offset(1, 1),
                      duration: 500.ms,
                      curve: Curves.elasticOut,
                    )
                    .fadeIn(duration: 300.ms),

                const SizedBox(height: 24),

                Text(
                  'Ты проголосовал!',
                  style: AppTextStyles.headline1,
                  textAlign: TextAlign.center,
                )
                    .animate(delay: 200.ms)
                    .fadeIn(duration: 400.ms)
                    .slideY(begin: 0.2),

                const SizedBox(height: 12),

                Text(
                  'Reveal в пятницу в 20:00',
                  style: AppTextStyles.bodySecondary.copyWith(fontSize: 18),
                  textAlign: TextAlign.center,
                )
                    .animate(delay: 400.ms)
                    .fadeIn(duration: 400.ms),

                const SizedBox(height: 32),

                progressAsync.when(
                  loading: () => const CircularProgressIndicator(),
                  error: (_, _) => const SizedBox.shrink(),
                  data: (progress) => _buildProgressCard(progress),
                ),

                const Spacer(),

                SizedBox(
                  width: double.infinity,
                  height: 52,
                  child: ElevatedButton(
                    onPressed: () {
                      HapticFeedback.mediumImpact();
                      context.go('/groups/${widget.groupId}');
                    },
                    child: const Text('Назад в группу'),
                  ),
                )
                    .animate(delay: 800.ms)
                    .fadeIn(duration: 400.ms),

                const SizedBox(height: 16),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildProgressCard(GroupVotingProgress progress) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Text(
            'Прогресс группы',
            style:
                AppTextStyles.body.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 12),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: progress.totalCount > 0
                  ? progress.votedCount / progress.totalCount
                  : 0,
              backgroundColor: AppColors.surface,
              color: AppColors.primary,
              minHeight: 8,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            '${progress.votedCount} из ${progress.totalCount} проголосовали',
            style: AppTextStyles.caption,
          ),
          if (progress.quorumReached) ...[
            const SizedBox(height: 8),
            Text(
              'Кворум достигнут!',
              style: AppTextStyles.caption.copyWith(
                color: AppColors.success,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ],
      ),
    )
        .animate(delay: 600.ms)
        .fadeIn(duration: 400.ms)
        .slideY(begin: 0.2);
  }
}
