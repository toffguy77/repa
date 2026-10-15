import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../domain/voting.dart';
import 'voting_notifier.dart';
import '../../../core/theme/app_tokens.dart';

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
                      duration: context.motion(MotionClass.emphasis),
                      curve: Curves.elasticOut,
                    )
                    .fadeIn(duration: context.motion(MotionClass.surface)),

                const SizedBox(height: 24),

                Text(
                  'Ты проголосовал!',
                  style: context.ts.heading1,
                  textAlign: TextAlign.center,
                )
                    .animate(delay: AppTokens.motion.stagger(3))
                    .fadeIn(duration: context.motion(MotionClass.emphasis))
                    .slideY(begin: 0.2),

                const SizedBox(height: 12),

                Text(
                  'Reveal в пятницу в 20:00',
                  style: context.ts.bodySecondary.copyWith(fontSize: 18),
                  textAlign: TextAlign.center,
                )
                    .animate(delay: AppTokens.motion.stagger(7))
                    .fadeIn(duration: context.motion(MotionClass.emphasis)),

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
                    .animate(delay: AppTokens.motion.stagger(13))
                    .fadeIn(duration: context.motion(MotionClass.emphasis)),

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
        color: context.t.elevation.level1.surface,
        borderRadius: AppTokens.radius.card,
        border: Border.all(
          color: context.t.elevation.level1.outline,
          width: AppTokens.border.hairline,
        ),
        boxShadow: context.t.elevation.level1.shadows,
      ),
      child: Column(
        children: [
          Text(
            'Прогресс группы',
            style:
                context.ts.body.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 12),
          ClipRRect(
            borderRadius: BorderRadius.circular(AppTokens.radius.xs),
            child: LinearProgressIndicator(
              value: progress.totalCount > 0
                  ? progress.votedCount / progress.totalCount
                  : 0,
              backgroundColor: context.t.color.surface,
              color: context.t.color.accent,
              minHeight: 8,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            '${progress.votedCount} из ${progress.totalCount} проголосовали',
            style: context.ts.caption,
          ),
          if (progress.quorumReached) ...[
            const SizedBox(height: 8),
            Text(
              'Кворум достигнут!',
              style: context.ts.caption.copyWith(
                color: context.t.color.success,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ],
      ),
    )
        .animate(delay: AppTokens.motion.stagger(10))
        .fadeIn(duration: context.motion(MotionClass.emphasis))
        .slideY(begin: 0.2);
  }
}
