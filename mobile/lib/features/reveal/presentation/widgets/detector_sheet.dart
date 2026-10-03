import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../../groups/presentation/widgets/member_avatar.dart';
import '../../domain/reveal.dart';

/// The detector, as a ladder rather than a single purchase.
///
/// The free rung — how many people voted — is always visible, because its job is to *create* the
/// question. Charging for the question is how a single-purchase detector ends up selling only the
/// answer. See docs/features/reveal.md.
class DetectorSheet extends StatelessWidget {
  final DetectorResult? detector;
  final bool buying;
  final int crystalBalance;
  final VoidCallback onBuy;
  final VoidCallback onGoToShop;

  /// Buys one partial reveal. Optional so older call sites keep compiling.
  final VoidCallback? onBuyHint;

  const DetectorSheet({
    super.key,
    required this.detector,
    required this.buying,
    this.crystalBalance = 0,
    required this.onBuy,
    required this.onGoToShop,
    this.onBuyHint,
  });

  @override
  Widget build(BuildContext context) {
    final d = detector;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text('Кто голосовал за тебя', style: context.ts.heading2),
        SizedBox(height: AppTokens.space.sm),
        Text(
          'Детектор показывает, кто участвовал в голосовании, но не как именно ответил',
          style: context.ts.caption,
          textAlign: TextAlign.center,
        ),
        SizedBox(height: AppTokens.space.xl),

        // Rung 1 — free. Always visible, names nobody.
        AppStat(
          value: '${d?.voterCount ?? 0}',
          label: 'проголосовали про тебя',
          tone: AppStatTone.accent,
        ),
        SizedBox(height: AppTokens.space.xl),

        if (d != null && !d.available)
          const _TooSmallNotice()
        else if (d != null && d.purchased)
          ..._fullList(context, d)
        else if (d != null)
          ..._ladder(context, d),

        SizedBox(height: AppTokens.space.lg),
      ],
    );
  }

  /// The owned top rung: everyone, by name.
  List<Widget> _fullList(BuildContext context, DetectorResult d) => [
        for (final v in d.voters) _VoterTile(voter: v),
      ];

  /// The rungs still for sale, plus whatever hints have already been revealed.
  List<Widget> _ladder(BuildContext context, DetectorResult d) {
    final canAffordHint = crystalBalance >= d.hintCost;
    final canAffordFull = crystalBalance >= d.fullCost;

    return [
      for (final hint in d.hints) _HintTile(hint: hint),
      if (d.hints.isNotEmpty) SizedBox(height: AppTokens.space.md),

      if (d.hintAvailable)
        AppButton(
          label: canAffordHint
              ? 'Подсказка — ${d.hintCost} 💎'
              : 'Не хватает кристаллов на подсказку',
          variant: AppButtonVariant.secondary,
          loading: buying,
          onPressed: canAffordHint ? onBuyHint : onGoToShop,
        ),
      if (d.hintAvailable) SizedBox(height: AppTokens.space.sm),

      AppButton(
        label: canAffordFull
            ? 'Показать всех — ${d.fullCost} 💎'
            : 'Купить кристаллы',
        loading: buying,
        onPressed: canAffordFull ? onBuy : onGoToShop,
      ),
    ];
  }
}

/// A revealed voter: avatar and one character. Enough to guess, not enough to know.
class _HintTile extends StatelessWidget {
  const _HintTile({required this.hint});

  final DetectorHint hint;

  @override
  Widget build(BuildContext context) {
    return AppListTile(
      leading: MemberAvatar(
        avatarEmoji: hint.avatarEmoji,
        avatarUrl: hint.avatarUrl,
      ),
      title: '${hint.firstLetter}•••',
      subtitle: 'Подсказка',
    ).animate().fadeIn(duration: context.motion(MotionClass.surface));
  }
}

class _VoterTile extends StatelessWidget {
  const _VoterTile({required this.voter});

  final VoterProfile voter;

  @override
  Widget build(BuildContext context) {
    return AppListTile(
      leading: MemberAvatar(
        avatarEmoji: voter.avatarEmoji,
        avatarUrl: voter.avatarUrl,
      ),
      title: voter.username,
    ).animate().fadeIn(duration: context.motion(MotionClass.surface));
  }
}

class _TooSmallNotice extends StatelessWidget {
  const _TooSmallNotice();

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        DecoratedBox(
          decoration: BoxDecoration(
            color: context.t.color.warning.withValues(alpha: 0.1),
            borderRadius: AppTokens.radius.tile,
          ),
          child: Padding(
            padding: EdgeInsets.all(AppTokens.space.md),
            child: Text(
              'Детектор работает в группах от 5 человек — иначе и так понятно, '
              'кто голосовал. Позови ещё людей.',
              style: context.ts.caption,
              textAlign: TextAlign.center,
            ),
          ),
        ),
        SizedBox(height: AppTokens.space.md),
        const AppButton(label: 'Детектор недоступен', onPressed: null),
      ],
    );
  }
}
