import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../groups/presentation/widgets/member_avatar.dart';
import '../domain/reveal.dart';
import 'reveal_notifier.dart';
import 'widgets/attribute_bar.dart';

class MembersRevealScreen extends ConsumerStatefulWidget {
  final String groupId;
  final String seasonId;
  final String seasonStatus;

  const MembersRevealScreen({
    super.key,
    required this.groupId,
    required this.seasonId,
    this.seasonStatus = 'REVEALED',
  });

  @override
  ConsumerState<MembersRevealScreen> createState() =>
      _MembersRevealScreenState();
}

class _MembersRevealScreenState extends ConsumerState<MembersRevealScreen> {
  late final _args =
      (seasonId: widget.seasonId, status: widget.seasonStatus);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(revealProvider(_args).notifier).loadMembersCards();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(revealProvider(_args));
    final cards = state.membersCards;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Участники'),
      ),
      body: cards == null
          ? const Center(child: CircularProgressIndicator())
          : cards.isEmpty
              ? Center(
                  child: Text(
                    'Нет данных',
                    style: AppTextStyles.bodySecondary,
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: cards.length,
                  itemBuilder: (context, index) {
                    return _MemberCardTile(
                      card: cards[index],
                      index: index,
                      groupId: widget.groupId,
                    );
                  },
                ),
    );
  }
}

class _MemberCardTile extends StatelessWidget {
  final MemberCard card;
  final int index;
  final String groupId;

  const _MemberCardTile({
    required this.card,
    required this.index,
    required this.groupId,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {
        HapticFeedback.lightImpact();
        context.push('/groups/$groupId/members/${card.userId}');
      },
      child: Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
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
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              MemberAvatar(
                avatarEmoji: card.avatarEmoji,
                avatarUrl: card.avatarUrl,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(card.username, style: AppTextStyles.body.copyWith(
                      fontWeight: FontWeight.w600,
                    )),
                    Text(
                      card.reputationTitle,
                      style: AppTextStyles.caption.copyWith(
                        color: AppColors.primary,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          if (card.topAttributes.isNotEmpty) ...[
            const SizedBox(height: 16),
            ...List.generate(card.topAttributes.length, (i) {
              final attr = card.topAttributes[i];
              return Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: AttributeBar(
                  questionText: attr.questionText,
                  percentage: attr.percentage,
                  index: i,
                ),
              );
            }),
          ],
        ],
      ),
    ).animate()
        .fadeIn(duration: 300.ms, delay: Duration(milliseconds: index * 100))
        .slideY(begin: 0.1, duration: 300.ms),
    );
  }
}
