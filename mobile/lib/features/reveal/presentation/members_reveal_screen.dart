import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/providers/connectivity_provider.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../../core/widgets/empty_state_widget.dart';
import '../../../core/widgets/error_state_widget.dart';
import '../../../core/widgets/skeleton_loader.dart';
import '../../groups/presentation/widgets/member_avatar.dart';
import '../domain/reveal.dart';
import 'reveal_notifier.dart';
import 'widgets/attribute_bar.dart';
import 'widgets/reaction_bar.dart';

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

    // Auto-refresh on reconnect
    ref.listen<bool>(connectivityProvider, (prev, next) {
      if (prev == false && next == true) {
        ref.read(revealProvider(_args).notifier).loadMembersCards();
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text('Участники'),
      ),
      body: RefreshIndicator(
        onRefresh: () =>
            ref.read(revealProvider(_args).notifier).loadMembersCards(),
        child: _buildBody(state),
      ),
    );
  }

  Widget _buildBody(RevealState state) {
    final cards = state.membersCards;

    if (state.error != null && cards == null) {
      return ListView(
        children: [
          const SizedBox(height: 120),
          ErrorStateWidget(
            message: state.error,
            onRetry: () =>
                ref.read(revealProvider(_args).notifier).loadMembersCards(),
          ),
        ],
      );
    }

    if (cards == null) {
      return ListView(
        padding: const EdgeInsets.all(16),
        children: List.generate(3, (_) => const MemberCardSkeleton()),
      );
    }

    if (cards.isEmpty) {
      return ListView(
        children: const [
          SizedBox(height: 120),
          EmptyStateWidget(
            emoji: '\u{1F50D}',
            title: 'Нет данных',
            subtitle: 'Карточки участников пока недоступны',
          ),
        ],
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: cards.length,
      itemBuilder: (context, index) {
        return _MemberCardTile(
          card: cards[index],
          index: index,
          groupId: widget.groupId,
          seasonId: widget.seasonId,
          seasonStatus: widget.seasonStatus,
        );
      },
    );
  }
}

class _MemberCardTile extends ConsumerStatefulWidget {
  final MemberCard card;
  final int index;
  final String groupId;
  final String seasonId;
  final String seasonStatus;

  const _MemberCardTile({
    required this.card,
    required this.index,
    required this.groupId,
    required this.seasonId,
    required this.seasonStatus,
  });

  @override
  ConsumerState<_MemberCardTile> createState() => _MemberCardTileState();
}

class _MemberCardTileState extends ConsumerState<_MemberCardTile> {
  late final _args =
      (seasonId: widget.seasonId, status: widget.seasonStatus);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(revealProvider(_args).notifier).loadReactions(widget.card.userId);
    });
  }

  @override
  Widget build(BuildContext context) {
    final reactions = ref.watch(
      revealProvider(_args).select((s) => s.reactions[widget.card.userId]),
    );
    final card = widget.card;

    return GestureDetector(
      onTap: () {
        HapticFeedback.lightImpact();
        context.push('/groups/${widget.groupId}/members/${card.userId}');
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
          const SizedBox(height: 12),
          ReactionBar(
            counts: reactions?.counts ?? {},
            myEmoji: reactions?.myEmoji,
            onReact: (emoji) {
              ref.read(revealProvider(_args).notifier)
                  .sendReaction(card.userId, emoji);
            },
          ),
        ],
      ),
    ).animate()
        .fadeIn(duration: 300.ms, delay: Duration(milliseconds: widget.index * 100))
        .slideY(begin: 0.1, duration: 300.ms),
    );
  }
}
