import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';

/// Pulls the bare code out of whatever the user typed or pasted.
///
/// This is **not** a re-implementation of the backend's validity rule — the server still
/// normalises case and separators authoritatively. The client has to do this much because the
/// code travels as a URL *path segment*: a pasted link's slashes would break routing before
/// the request ever reached the service.
String extractInviteCode(String raw) {
  var s = raw.trim();
  if (s.contains('/')) {
    s = s.split('/').last;
  }
  final cut = s.indexOf(RegExp(r'[?#]'));
  if (cut >= 0) {
    s = s.substring(0, cut);
  }
  return s.trim();
}

/// Groups a 6-character invite code so it reads as two chunks instead of one blur.
///
/// Mirrors `FormatInviteCode` on the backend. The gap is cosmetic — the server strips any
/// separator back out, so a user who types what they see succeeds.
String formatInviteCode(String code) =>
    code.length == 6 ? '${code.substring(0, 3)} ${code.substring(3)}' : code;

/// The invite URL for a code. Must match `InviteURL` on the backend.
String inviteUrlFor(String code) => 'https://repa.app/join/$code';

/// The sheet shown after creating a group and from the group's invite action.
///
/// Shows the **code** as prominently as the link: the audience sits in the same room as the
/// people it needs to invite, so "код AB2 CD3, вбей" is a real distribution path that a
/// 36-character link cannot serve.
class InviteShareSheet extends StatelessWidget {
  const InviteShareSheet({
    super.key,
    required this.inviteCode,
    this.title = 'Позови своих',
    this.emoji = '\u{1F389}',
    this.onDone,
  });

  final String inviteCode;
  final String title;
  final String emoji;
  final VoidCallback? onDone;

  String get _url => inviteUrlFor(inviteCode);

  void _copy(BuildContext context, String value, String message) {
    Clipboard.setData(ClipboardData(text: value));
    HapticFeedback.lightImpact();
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    final t = context.t;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(emoji, style: const TextStyle(fontSize: 48)),
        SizedBox(height: AppTokens.space.md),
        Text(title, style: context.ts.heading2),
        SizedBox(height: AppTokens.space.sm),
        Text(
          'Скинь ссылку или продиктуй код',
          style: context.ts.bodySecondary,
          textAlign: TextAlign.center,
        ),
        SizedBox(height: AppTokens.space.xl),

        // The code, big enough to read across a room.
        GestureDetector(
          onTap: () => _copy(context, inviteCode, 'Код скопирован'),
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: t.color.accentFill.withValues(alpha: 0.14),
              borderRadius: AppTokens.radius.card,
              border: Border.all(
                color: t.color.accent.withValues(alpha: 0.35),
                width: AppTokens.border.hairline,
              ),
            ),
            child: Padding(
              padding: EdgeInsets.symmetric(
                horizontal: AppTokens.space.xl,
                vertical: AppTokens.space.lg,
              ),
              child: Column(
                children: [
                  Text(
                    'КОД ГРУППЫ',
                    style: AppTokens.text.label.copyWith(color: t.color.accent),
                  ),
                  SizedBox(height: AppTokens.space.xs),
                  Text(
                    formatInviteCode(inviteCode),
                    style: AppTokens.text.display2.copyWith(
                      color: t.color.textPrimary,
                      fontFeatures: AppTokens.text.numeric.fontFeatures,
                      letterSpacing: 2,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        SizedBox(height: AppTokens.space.md),

        // The link, copyable independently of the code.
        DecoratedBox(
          decoration: BoxDecoration(
            color: t.color.surface,
            borderRadius: AppTokens.radius.tile,
          ),
          child: Padding(
            padding: EdgeInsets.only(left: AppTokens.space.lg),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    _url,
                    style: context.ts.caption,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.copy, size: 20),
                  tooltip: 'Скопировать ссылку',
                  onPressed: () => _copy(context, _url, 'Ссылка скопирована'),
                ),
              ],
            ),
          ),
        ),
        SizedBox(height: AppTokens.space.lg),

        AppButton(
          label: 'Поделиться',
          icon: Icons.share,
          onPressed: () => Share.share(_url),
        ),
        if (onDone != null) ...[
          SizedBox(height: AppTokens.space.sm),
          AppButton(
            label: 'Готово',
            variant: AppButtonVariant.ghost,
            onPressed: onDone,
          ),
        ],
      ],
    );
  }
}
