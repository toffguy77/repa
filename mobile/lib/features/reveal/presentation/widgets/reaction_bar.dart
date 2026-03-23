import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';

const reactionEmojis = ['😂', '🔥', '💀', '👀', '🫡'];

class ReactionBar extends StatelessWidget {
  final Map<String, int> counts;
  final String? myEmoji;
  final ValueChanged<String> onReact;

  const ReactionBar({
    super.key,
    required this.counts,
    this.myEmoji,
    required this.onReact,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: reactionEmojis.map((emoji) {
        final count = counts[emoji] ?? 0;
        final isSelected = myEmoji == emoji;

        return Padding(
          padding: const EdgeInsets.only(right: 6),
          child: Semantics(
            label: 'Реакция $emoji${count > 0 ? ", $count" : ""}${isSelected ? ", выбрано" : ""}',
            button: true,
            child: GestureDetector(
            onTap: () {
              HapticFeedback.lightImpact();
              onReact(emoji);
            },
            child: AnimatedContainer(
              duration: 200.ms,
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: isSelected
                    ? AppColors.primaryLight
                    : Colors.grey.shade100,
                borderRadius: BorderRadius.circular(20),
                border: isSelected
                    ? Border.all(color: AppColors.primary, width: 1.5)
                    : null,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(emoji, style: const TextStyle(fontSize: 16)),
                  if (count > 0) ...[
                    const SizedBox(width: 4),
                    Text(
                      '$count',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: isSelected
                            ? AppColors.primary
                            : AppColors.textSecondary,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
          ),
        );
      }).toList(),
    );
  }
}
