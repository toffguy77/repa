import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../domain/reveal.dart';
import '../../../groups/presentation/widgets/member_avatar.dart';

class DetectorSheet extends StatelessWidget {
  final DetectorResult? detector;
  final bool buying;
  final int crystalBalance;
  final VoidCallback onBuy;
  final VoidCallback onGoToShop;

  const DetectorSheet({
    super.key,
    required this.detector,
    required this.buying,
    this.crystalBalance = 0,
    required this.onBuy,
    required this.onGoToShop,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(height: 20),
          Text(
            'Кто голосовал за тебя',
            style: AppTextStyles.headline2,
          ),
          const SizedBox(height: 8),
          Text(
            'Детектор показывает кто участвовал в голосовании, но не как именно ответил',
            style: AppTextStyles.caption,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 20),
          if (detector != null && detector!.purchased) ...[
            ...detector!.voters.map((v) => _buildVoterTile(v)),
          ] else ...[
            // Blurred placeholder
            ...List.generate(4, (i) => _buildBlurredTile(i)),
            const SizedBox(height: 16),
            if (crystalBalance < 10)
              Column(
                children: [
                  Text(
                    'Не хватает кристаллов (\u{1F48E} $crystalBalance / 10)',
                    style: AppTextStyles.caption,
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    width: double.infinity,
                    height: 52,
                    child: ElevatedButton(
                      onPressed: () {
                        HapticFeedback.mediumImpact();
                        Navigator.of(context).pop();
                        onGoToShop();
                      },
                      child: const Text('Купить кристаллы'),
                    ),
                  ),
                ],
              )
            else
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton.icon(
                  onPressed: buying
                      ? null
                      : () {
                          HapticFeedback.mediumImpact();
                          onBuy();
                        },
                  icon: buying
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Text('\u{1F48E}'),
                  label: Text(buying ? 'Покупка...' : 'Узнать за 10'),
                ),
              ),
          ],
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Widget _buildVoterTile(VoterProfile voter) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          MemberAvatar(
            avatarEmoji: voter.avatarEmoji,
            avatarUrl: voter.avatarUrl,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(voter.username, style: AppTextStyles.body),
          ),
        ],
      ).animate().fadeIn(duration: 300.ms).slideX(begin: 0.05),
    );
  }

  Widget _buildBlurredTile(int index) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: AppColors.surface,
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 12),
          Container(
            width: 80 + index * 20.0,
            height: 14,
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(7),
            ),
          ),
        ],
      ),
    );
  }
}
