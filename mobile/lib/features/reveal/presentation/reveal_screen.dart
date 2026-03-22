import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:share_plus/share_plus.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../domain/reveal.dart';
import 'reveal_notifier.dart';
import 'widgets/achievement_popup.dart';
import 'widgets/detector_sheet.dart';
import 'widgets/reputation_card.dart';

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
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) => Consumer(
        builder: (context, ref, _) {
          final state = ref.watch(revealProvider(_args));
          return DetectorSheet(
            detector: state.detector,
            buying: state.buyingDetector,
            onBuy: () =>
                ref.read(revealProvider(_args).notifier).buyDetector(),
          );
        },
      ),
    );
  }

  void _shareCard() {
    final state = ref.read(revealProvider(_args));
    final url = state.cardImageUrl;
    if (url != null) {
      Share.share('$url\n\nМоя репа этой недели!');
    } else {
      Share.share('Смотри мою репу! repa.app');
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
      backgroundColor: state.phase == RevealPhase.opening
          ? const Color(0xFF1a0d2e)
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
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('\u{1F346}', style: TextStyle(fontSize: 64)),
            const SizedBox(height: 24),
            Text(
              'Результаты ещё не готовы',
              style: AppTextStyles.headline2,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            Text(
              'Голосование завершится в пятницу в 20:00',
              style: AppTextStyles.bodySecondary,
              textAlign: TextAlign.center,
            ),
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
                  duration: 800.ms,
                ),
            const SizedBox(height: 32),
            Text(
              'Твоя репа готова!',
              style: AppTextStyles.headline1,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: _startOpening,
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.primary,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
                child: const Text(
                  'Открыть репу',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700),
                ),
              ),
            ),
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
                duration: 600.ms,
              )
              .then()
              .fadeOut(duration: 500.ms, delay: 1500.ms),
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
                duration: 600.ms,
                curve: Curves.easeOutCubic,
              )
              .fadeIn(duration: 400.ms),

          const SizedBox(height: 24),

          // Action buttons
          _buildActionButtons(),

          const SizedBox(height: 24),

          // Members cards button
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: _openMembersCards,
              icon: const Icon(Icons.people_outline),
              label: const Text('Карточки участников'),
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.primary,
                side: const BorderSide(color: AppColors.primary),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                padding: const EdgeInsets.symmetric(vertical: 14),
              ),
            ),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _buildUserHeader(RevealData data) {
    return Column(
      children: [
        Container(
          width: 80,
          height: 80,
          decoration: BoxDecoration(
            color: AppColors.primaryLight,
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              '\u{1F346}',
              style: const TextStyle(fontSize: 40),
            ),
          ),
        ),
        const SizedBox(height: 12),
        Text(
          data.myCard.reputationTitle,
          style: AppTextStyles.headline2.copyWith(
            color: AppColors.primary,
          ),
        ),
      ],
    );
  }

  Widget _buildActionButtons() {
    return Row(
      children: [
        Expanded(
          child: ElevatedButton.icon(
            onPressed: _shareCard,
            icon: const Icon(Icons.share, size: 20),
            label: const Text('Поделиться'),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              padding: const EdgeInsets.symmetric(vertical: 14),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: OutlinedButton.icon(
            onPressed: _showDetector,
            icon: const Text('\u{1F50D}'),
            label: const Text('Детектор'),
            style: OutlinedButton.styleFrom(
              foregroundColor: AppColors.primary,
              side: const BorderSide(color: AppColors.primary),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              padding: const EdgeInsets.symmetric(vertical: 14),
            ),
          ),
        ),
      ],
    );
  }
}
