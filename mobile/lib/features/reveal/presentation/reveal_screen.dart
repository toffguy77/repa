import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import '../../../core/providers/auth_provider.dart';
import '../../crystals/presentation/crystals_notifier.dart';
import '../../../core/widgets/reveal_countdown_widget.dart';
import '../../groups/domain/group.dart';
import '../../groups/presentation/groups_notifier.dart';
import '../../telegram/presentation/telegram_notifier.dart';
import '../domain/reveal.dart';
import '../../../core/analytics/analytics_service.dart';
import 'reveal_notifier.dart';
import 'widgets/achievement_popup.dart';
import 'widgets/detector_sheet.dart';
import 'widgets/reputation_card.dart';
import '../../../core/theme/app_tokens.dart';
import '../../../core/theme/brand_colors.dart';
import '../../../core/widgets/kit/kit.dart';
import '../domain/share_link.dart';
import 'widgets/anticipation_panel.dart';

class RevealScreen extends ConsumerStatefulWidget {
  final String groupId;
  final String seasonId;
  final String seasonStatus;

  const RevealScreen({
    super.key,
    required this.groupId,
    required this.seasonId,
    required this.seasonStatus,
  });

  @override
  ConsumerState<RevealScreen> createState() => _RevealScreenState();
}

class _RevealScreenState extends ConsumerState<RevealScreen> {
  bool _showAchievements = false;
  late final _args =
      (seasonId: widget.seasonId, status: widget.seasonStatus);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(analyticsProvider).logRevealOpened(widget.groupId);
      ref.read(revealProvider(_args).notifier).load();
    });
  }

  void _startOpening() {
    HapticFeedback.heavyImpact();
    ref.read(revealProvider(_args).notifier).startOpening();

    // After 3 seconds, finish opening
    Future.delayed(const Duration(seconds: 3), () {
      if (mounted) {
        ref.read(revealProvider(_args).notifier).finishOpening();
        // Show achievements if any
        final state = ref.read(revealProvider(_args));
        if (state.data != null &&
            state.data!.myCard.newAchievements.isNotEmpty) {
          setState(() => _showAchievements = true);
        }
      }
    });
  }

  void _showDetector() {
    ref.read(revealProvider(_args).notifier).loadDetector();
    ref.read(crystalBalanceProvider.notifier).load();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) => Consumer(
        builder: (context, ref, _) {
          final state = ref.watch(revealProvider(_args));
          final balance = ref.watch(crystalBalanceProvider);
          return DetectorSheet(
            detector: state.detector,
            buying: state.buyingDetector,
            crystalBalance: balance,
            onBuy: () =>
                ref.read(revealProvider(_args).notifier).buyDetector(),
            onBuyHint: () =>
                ref.read(revealProvider(_args).notifier).buyDetectorHint(),
            onGoToShop: () => context.push('/shop'),
          );
        },
      ),
    );
  }

  /// The group's invite code, so a shared card carries a way back into the product.
  String? get _inviteCode =>
      ref.read(groupDetailProvider(widget.groupId)).detail?.group.inviteCode;

  /// Reports a share for the funnel. Never surfaced: the share has already happened in the OS
  /// share sheet by the time this runs, so an error here would describe a failure the user did
  /// not have.
  void _reportShare(ShareChannel channel) {
    ref
        .read(revealRepositoryProvider)
        .recordShare(widget.seasonId, channel.wireValue)
        .catchError((Object e) {
      debugPrint('failed to record share: $e');
    });
  }

  Future<void> _shareCard() async {
    final state = ref.read(revealProvider(_args));
    final imageUrl = state.data?.myCard.cardImageUrl ?? state.cardImageUrl;
    final code = _inviteCode;

    // Without a code there is nothing to invite anyone to; the card still shares, just as a
    // picture rather than an invitation.
    final text = code == null
        ? 'Моя репа 🍆 repa.app'
        : buildShareText(
            inviteCode: code,
            channel: ShareChannel.card,
            referrerId: ref.read(authProvider).user?.id,
          );

    if (imageUrl != null) {
      try {
        final dir = await getTemporaryDirectory();
        final file = File('${dir.path}/repa_card_${widget.seasonId}.png');
        await Dio().download(imageUrl, file.path);
        await Share.shareXFiles([XFile(file.path)], text: text);
        _reportShare(ShareChannel.card);
        return;
      } catch (_) {
        // Fall through to a text-only share.
      }
    }

    await Share.share(text);
    _reportShare(ShareChannel.card);
  }

  Future<void> _shareToTelegram() async {
    HapticFeedback.mediumImpact();
    try {
      await ref
          .read(shareToTelegramProvider)
          .shareToTelegram(widget.seasonId);
      _reportShare(ShareChannel.telegram);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Карточка опубликована в чате')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString())),
        );
      }
    }
  }

  void _openMembersCards() {
    context.push(
        '/groups/${widget.groupId}/reveal/${widget.seasonId}/members?status=${widget.seasonStatus}');
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(revealProvider(_args));

    ref.listen<RevealState>(revealProvider(_args), (prev, next) {
      if (next.error != null && prev?.error != next.error) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(next.error!)),
        );
        ref.read(revealProvider(_args).notifier).clearError();
      }
    });

    return Scaffold(
      // The opening animation is an always-dark brand moment, matching the shared card.
      backgroundColor: state.phase == RevealPhase.opening
          ? AppColorTokens.dark.canvas
          : null,
      appBar: state.phase == RevealPhase.opening
          ? null
          : AppBar(
              title: const Text('Reveal'),
            ),
      body: Stack(
        children: [
          _buildBody(state),
          if (_showAchievements &&
              state.data != null &&
              state.data!.myCard.newAchievements.isNotEmpty)
            AchievementPopup(
              achievements: state.data!.myCard.newAchievements,
              onDismiss: () => setState(() => _showAchievements = false),
            ),
        ],
      ),
    );
  }

  Widget _buildBody(RevealState state) {
    switch (state.phase) {
      case RevealPhase.loading:
        return const Center(child: CircularProgressIndicator());

      case RevealPhase.waiting:
        return _buildWaiting();

      case RevealPhase.ready:
        return _buildReady();

      case RevealPhase.opening:
        return _buildOpening();

      case RevealPhase.revealed:
        return _buildRevealed(state);
    }
  }

  Widget _buildWaiting() {
    final revealState = ref.watch(revealProvider(_args));
    // The group's active season knows *why* the Reveal has not happened — waiting for
    // members, waiting for voters, postponed, or simply not Friday yet. Showing the same
    // explanation here is what keeps the Tuesday/Thursday pushes honest.
    final season =
        ref.watch(groupDetailProvider(widget.groupId)).detail?.activeSeason;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('\u{1F346}', style: TextStyle(fontSize: 64)),
            const SizedBox(height: 24),
            Text(
              season?.revealHeadline ?? 'Результаты ещё не готовы',
              style: context.ts.heading2,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            Text(
              season?.revealExplanation ??
                  'Голосование завершится в пятницу в 20:00',
              style: context.ts.bodySecondary,
              textAlign: TextAlign.center,
            ),
            // The anticipation panel carries its own countdown, so the standalone one is only for
            // the case where there is nothing else to say.
            if (revealState.anticipation != null) ...[
              SizedBox(height: AppTokens.space.xl),
              AnticipationPanel(state: revealState.anticipation),
            ] else if (season != null &&
                season.revealState == SeasonRevealState.scheduled) ...[
              const SizedBox(height: 24),
              RevealCountdownWidget(revealAt: season.revealAt),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildReady() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('\u{1F346}', style: TextStyle(fontSize: 80))
                .animate(onPlay: (c) => c.repeat(reverse: true))
                .scale(
                  begin: const Offset(1, 1),
                  end: const Offset(1.15, 1.15),
                  duration: context.motion(MotionClass.emphasis),
                ),
            const SizedBox(height: 32),
            Text(
              'Твоя репа готова!',
              style: context.ts.heading1,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            AppButton(label: 'Открыть репу', onPressed: _startOpening),
          ],
        ),
      ),
    );
  }

  Widget _buildOpening() {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text('\u{1F346}', style: TextStyle(fontSize: 100))
              .animate(onPlay: (c) => c.repeat(reverse: true))
              .scale(
                begin: const Offset(0.8, 0.8),
                end: const Offset(1.3, 1.3),
                duration: context.motion(MotionClass.emphasis),
              )
              .then()
              .fadeOut(duration: context.motion(MotionClass.emphasis), delay: AppTokens.motion.stagger(25)),
        ],
      ),
    );
  }

  Widget _buildRevealed(RevealState state) {
    final data = state.data;
    if (data == null) return const SizedBox.shrink();

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          // Avatar + username header
          _buildUserHeader(data),
          const SizedBox(height: 16),

          // Reputation card
          ReputationCard(
            card: data.myCard,
            onOpenHidden: () =>
                ref.read(revealProvider(_args).notifier).openHidden(),
            unlockingHidden: state.unlockingHidden,
          ).animate()
              .slideY(
                begin: 1,
                end: 0,
                duration: context.motion(MotionClass.emphasis),
                curve: Curves.easeOutCubic,
              )
              .fadeIn(duration: context.motion(MotionClass.emphasis)),

          const SizedBox(height: 24),

          // Action buttons
          _buildActionButtons(),

          const SizedBox(height: 24),

          // Members cards button
          AppButton(
            label: 'Карточки участников',
            icon: Icons.people_outline,
            variant: AppButtonVariant.secondary,
            onPressed: _openMembersCards,
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _buildUserHeader(RevealData data) {
    final user = ref.watch(authProvider).user;
    final emoji = user?.avatarEmoji ?? '\u{1F346}';

    return Column(
      children: [
        Container(
          width: 80,
          height: 80,
          decoration: BoxDecoration(
            color: context.t.color.accentFill.withValues(alpha: 0.18),
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              emoji,
              style: const TextStyle(fontSize: 40),
            ),
          ),
        ),
        const SizedBox(height: 12),
        Text(
          data.myCard.reputationTitle,
          style: context.ts.heading2.copyWith(
            color: context.t.color.accent,
          ),
        ),
      ],
    );
  }

  Widget _buildActionButtons() {
    final groupState = ref.watch(groupDetailProvider(widget.groupId));
    final hasTelegram =
        groupState.detail?.group.telegramUsername != null;

    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: AppButton(
                label: 'Поделиться',
                icon: Icons.share,
                onPressed: _shareCard,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: AppButton(
                label: 'Детектор',
                icon: Icons.search,
                variant: AppButtonVariant.secondary,
                onPressed: _showDetector,
              ),
            ),
          ],
        ),
        if (hasTelegram) ...[
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: _shareToTelegram,
              icon: const Icon(Icons.telegram, size: 20),
              label: const Text('Отправить в Telegram-чат'),
              style: OutlinedButton.styleFrom(
                foregroundColor: BrandColors.telegram,
                side: const BorderSide(color: BrandColors.telegram),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(AppTokens.radius.md),
                ),
                padding: const EdgeInsets.symmetric(vertical: 14),
              ),
            ),
          ),
        ],
      ],
    );
  }
}
