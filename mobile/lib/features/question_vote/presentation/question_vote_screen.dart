import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../../core/widgets/error_state_widget.dart';
import '../../../core/widgets/skeleton_loader.dart';
import 'question_vote_notifier.dart';
import 'widgets/question_candidate_card.dart';

class QuestionVoteScreen extends ConsumerStatefulWidget {
  final String groupId;

  const QuestionVoteScreen({super.key, required this.groupId});

  @override
  ConsumerState<QuestionVoteScreen> createState() => _QuestionVoteScreenState();
}

class _QuestionVoteScreenState extends ConsumerState<QuestionVoteScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(questionVoteProvider(widget.groupId).notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(questionVoteProvider(widget.groupId));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Вопрос недели'),
      ),
      body: switch (state.status) {
        QuestionVoteStatus.loading => _buildLoading(),
        QuestionVoteStatus.voting => _buildVoting(state),
        QuestionVoteStatus.voted => _buildVoted(state),
        QuestionVoteStatus.unavailable => _buildUnavailable(),
        QuestionVoteStatus.error => ErrorStateWidget(
            message: state.error,
            onRetry: () => ref
                .read(questionVoteProvider(widget.groupId).notifier)
                .load(),
          ),
      },
    );
  }

  Widget _buildLoading() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const SkeletonLoader(height: 24, width: 250, borderRadius: 6),
          const SizedBox(height: 8),
          const SkeletonLoader(height: 16, width: 200, borderRadius: 6),
          const SizedBox(height: 24),
          ...List.generate(
            3,
            (_) => const Padding(
              padding: EdgeInsets.only(bottom: 12),
              child: SkeletonLoader(height: 80, borderRadius: 16),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildVoting(QuestionVoteState state) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          'Какой вопрос добавить на следующей неделе?',
          style: AppTextStyles.headline2,
        ),
        const SizedBox(height: 4),
        Text(
          'Победивший вопрос войдёт в ротацию',
          style: AppTextStyles.bodySecondary,
        ),
        const SizedBox(height: 24),
        ...state.candidates.asMap().entries.map(
              (entry) => Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: QuestionCandidateCard(
                  candidate: entry.value,
                  onTap: () => ref
                      .read(questionVoteProvider(widget.groupId).notifier)
                      .vote(entry.value.id),
                ).animate().fadeIn(
                      delay: Duration(milliseconds: 100 * entry.key),
                      duration: 300.ms,
                    ),
              ),
            ),
      ],
    );
  }

  Widget _buildVoted(QuestionVoteState state) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          'Какой вопрос добавить на следующей неделе?',
          style: AppTextStyles.headline2,
        ),
        const SizedBox(height: 4),
        Text(
          'Победивший вопрос войдёт в ротацию',
          style: AppTextStyles.bodySecondary,
        ),
        const SizedBox(height: 24),
        if (state.candidates.isNotEmpty)
          ...state.candidates.map(
            (c) => Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: QuestionCandidateCard(
                candidate: c,
                selected: c.id == state.selectedId,
                disabled: true,
              ),
            ),
          )
        else
          Center(
            child: Column(
              children: [
                const SizedBox(height: 32),
                const Text('\u{2705}', style: TextStyle(fontSize: 64)),
                const SizedBox(height: 16),
                Text(
                  'Готово!',
                  style: AppTextStyles.headline2,
                ),
                const SizedBox(height: 8),
                Text(
                  'Узнаешь результат в понедельник',
                  style: AppTextStyles.bodySecondary,
                ),
              ],
            ),
          ),
        if (state.candidates.isNotEmpty) ...[
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.primaryLight,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Row(
              children: [
                const Text('\u{2705}', style: TextStyle(fontSize: 24)),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Готово! Узнаешь результат в понедельник',
                    style: AppTextStyles.body.copyWith(
                      color: AppColors.primary,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
          ).animate().fadeIn(duration: 300.ms).slideY(
                begin: 0.3,
                end: 0,
                curve: Curves.easeOut,
              ),
        ],
      ],
    );
  }

  Widget _buildUnavailable() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('\u{1F5F3}', style: TextStyle(fontSize: 64)),
            const SizedBox(height: 24),
            Text(
              'Голосование за вопросы',
              style: AppTextStyles.headline2,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'Откроется в воскресенье в полдень',
              style: AppTextStyles.bodySecondary,
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}
