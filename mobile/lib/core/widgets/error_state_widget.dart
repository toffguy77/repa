import 'package:flutter/material.dart';

import '../theme/app_tokens.dart';
import 'kit/kit.dart';

class ErrorStateWidget extends StatelessWidget {
  final String? message;
  final VoidCallback? onRetry;

  const ErrorStateWidget({
    super.key,
    this.message,
    this.onRetry,
  });

  static String friendlyMessage(String? raw) {
    if (raw == null || raw.isEmpty) return 'Что-то пошло не так';
    final lower = raw.toLowerCase();
    if (lower.contains('connection') ||
        lower.contains('timeout') ||
        lower.contains('socket') ||
        lower.contains('network') ||
        lower.contains('соединен')) {
      return 'Нет соединения, проверь интернет';
    }
    if (lower.contains('500') || lower.contains('server') || lower.contains('internal')) {
      return 'Что-то пошло не так, попробуй позже';
    }
    return raw;
  }

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final displayMessage = friendlyMessage(message);

    return Center(
      child: Padding(
        padding: EdgeInsets.symmetric(horizontal: AppTokens.space.xxl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 64,
              height: 64,
              decoration: BoxDecoration(
                color: t.color.danger.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: Icon(
                Icons.error_outline_rounded,
                color: t.color.danger,
                size: 32,
              ),
            ),
            SizedBox(height: AppTokens.space.lg),
            Text(
              displayMessage,
              style: AppTokens.text.body.copyWith(color: t.color.textPrimary),
              textAlign: TextAlign.center,
            ),
            if (onRetry != null) ...[
              SizedBox(height: AppTokens.space.xl),
              AppButton(label: 'Повторить', onPressed: onRetry),
            ],
          ],
        ),
      ),
    );
  }
}
