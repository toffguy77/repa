import 'package:flutter/material.dart';

import '../../theme/app_tokens.dart';

/// A compact row. Dense by default, but never shorter than the minimum touch target.
class AppListTile extends StatelessWidget {
  const AppListTile({
    super.key,
    required this.title,
    this.subtitle,
    this.leading,
    this.trailing,
    this.onTap,
  });

  final String title;
  final String? subtitle;
  final Widget? leading;
  final Widget? trailing;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final radius = AppTokens.radius.tile;

    final row = ConstrainedBox(
      constraints: BoxConstraints(minHeight: AppTokens.border.minTouchTarget),
      child: Padding(
        padding: EdgeInsets.symmetric(
          horizontal: AppTokens.space.md,
          vertical: AppTokens.space.sm,
        ),
        child: Row(
          children: [
            if (leading != null) ...[
              leading!,
              SizedBox(width: AppTokens.space.md),
            ],
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: AppTokens.text.bodyStrong.copyWith(color: t.color.textPrimary),
                  ),
                  if (subtitle != null)
                    Text(
                      subtitle!,
                      style: AppTokens.text.caption.copyWith(color: t.color.textSecondary),
                    ),
                ],
              ),
            ),
            if (trailing != null) ...[
              SizedBox(width: AppTokens.space.md),
              trailing!,
            ],
          ],
        ),
      ),
    );

    if (onTap == null) return row;
    return Material(
      color: Colors.transparent,
      borderRadius: radius,
      child: InkWell(onTap: onTap, borderRadius: radius, child: row),
    );
  }
}
