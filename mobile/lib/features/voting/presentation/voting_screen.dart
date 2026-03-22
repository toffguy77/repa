import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../voting/domain/voting.dart';
import 'voting_notifier.dart';
import 'widgets/question_card.dart';
import 'widgets/participant_card.dart';

class VotingScreen extends ConsumerStatefulWidget {
  final String groupId;
  final String seasonId;

  const VotingScreen({
    super.key,
    required this.groupId,
    required this.seasonId,
  });

  @override
  ConsumerState<VotingScreen> createState() => _VotingScreenState();
}

class _VotingScreenState extends ConsumerState<VotingScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(votingProvider(widget.seasonId).notifier).loadSession();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(votingProvider(widget.seasonId));
    final notifier = ref.read(votingProvider(widget.seasonId).notifier);

    // Listen for errors
    ref.listen<VotingState>(votingProvider(widget.seasonId), (prev, next) {
      if (next.error != null && prev?.error != next.error) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.error!),
            backgroundColor: AppColors.error,
          ),
        );
        notifier.clearError();
      }
    });

    if (state.loading && state.session == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (state.error != null && state.session == null) {
      return Scaffold(
        appBar: AppBar(),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(state.error!, style: AppTextStyles.body),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => notifier.loadSession(),
                child: const Text('Повторить'),
              ),
            ],
          ),
        ),
      );
    }

    if (state.completed) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        context.go(
            '/groups/${widget.groupId}/vote/${widget.seasonId}/complete');
      });
      return const SizedBox.shrink();
    }

    final question = state.currentQuestion;
    if (question == null) return const SizedBox.shrink();

    final targets = notifier.shuffledTargets;
    final totalQuestions = state.totalQuestions;
    final answeredQuestions = state.answeredQuestions;

    return PopScope(
      canPop: false,
      child: Scaffold(
        appBar: AppBar(
          leading: IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => _showExitDialog(context),
          ),
          title: Text(
            '${answeredQuestions + 1} из $totalQuestions',
            style: AppTextStyles.body.copyWith(fontWeight: FontWeight.w600),
          ),
          centerTitle: true,
        ),
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Column(
              children: [
                // Progress bar
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: totalQuestions > 0
                        ? (answeredQuestions) / totalQuestions
                        : 0,
                    backgroundColor: AppColors.surface,
                    color: AppColors.primary,
                    minHeight: 6,
                  ),
                ),
                const SizedBox(height: 24),

                // Question card
                QuestionCard(
                  key: ValueKey(question.questionId),
                  text: question.text,
                  category: question.category,
                ),
                const SizedBox(height: 24),

                // Participants
                Expanded(
                  child: _buildTargets(targets, state),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTargets(List<VotingTarget> targets, VotingState state) {
    if (targets.length <= 4) {
      return GridView.count(
        crossAxisCount: 2,
        mainAxisSpacing: 12,
        crossAxisSpacing: 12,
        childAspectRatio: 1.0,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        children: targets.map((t) {
          return ParticipantCard(
            username: t.username,
            avatarEmoji: t.avatarEmoji,
            avatarUrl: t.avatarUrl,
            selected: state.selectedTargetId == t.userId,
            disabled: state.submitting,
            onTap: () => ref
                .read(votingProvider(widget.seasonId).notifier)
                .selectTarget(t.userId),
          );
        }).toList(),
      );
    }

    return GridView.builder(
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        mainAxisSpacing: 12,
        crossAxisSpacing: 12,
        childAspectRatio: 1.0,
      ),
      itemCount: targets.length,
      itemBuilder: (context, index) {
        final t = targets[index];
        return ParticipantCard(
          username: t.username,
          avatarEmoji: t.avatarEmoji,
          avatarUrl: t.avatarUrl,
          selected: state.selectedTargetId == t.userId,
          disabled: state.submitting,
          onTap: () => ref
              .read(votingProvider(widget.seasonId).notifier)
              .selectTarget(t.userId),
        );
      },
    );
  }

  void _showExitDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Выйти?'),
        content: const Text(
          'Прогресс сохранён. Вы сможете продолжить голосование позже.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Остаться'),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(ctx).pop();
              context.go('/groups/${widget.groupId}');
            },
            child: Text(
              'Выйти',
              style: TextStyle(color: AppColors.error),
            ),
          ),
        ],
      ),
    );
  }
}
