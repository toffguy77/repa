import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../groups/presentation/widgets/member_avatar.dart';
import '../../../../core/theme/app_tokens.dart';

class ParticipantCard extends StatelessWidget {
  final String username;
  final String? avatarEmoji;
  final String? avatarUrl;
  final bool selected;
  final bool disabled;
  final VoidCallback onTap;

  const ParticipantCard({
    super.key,
    required this.username,
    this.avatarEmoji,
    this.avatarUrl,
    this.selected = false,
    this.disabled = false,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: '$username${selected ? ", выбран" : ""}',
      button: true,
      child: GestureDetector(
      onTap: disabled
          ? null
          : () {
              HapticFeedback.mediumImpact();
              onTap();
            },
      child: AnimatedContainer(
        duration: context.motion(MotionClass.surface),
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 12),
        decoration: BoxDecoration(
          color: selected ? context.t.color.accentFill.withValues(alpha: 0.18) : context.t.elevation.level1.surface,
          borderRadius: BorderRadius.circular(AppTokens.radius.lg),
          border: Border.all(
            color: selected ? context.t.color.accent : Colors.grey.shade200,
            width: selected ? 2.5 : 1,
          ),
          boxShadow: [
            if (!selected)
              BoxShadow(
                color: context.t.color.scrim.withValues(alpha: 0.04),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
          ],
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Stack(
              alignment: Alignment.bottomRight,
              children: [
                MemberAvatar(
                  avatarEmoji: avatarEmoji,
                  avatarUrl: avatarUrl,
                  size: 52,
                ),
                if (selected)
                  Container(
                    width: 22,
                    height: 22,
                    decoration: BoxDecoration(
                      color: context.t.color.accentFill,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      Icons.check,
                      color: context.t.color.onAccentFill,
                      size: 14,
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              username,
              style: context.ts.caption.copyWith(
                fontWeight: selected ? FontWeight.w600 : FontWeight.normal,
                color: selected ? context.t.color.accent : context.t.color.textPrimary,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    ),
    );
  }
}
