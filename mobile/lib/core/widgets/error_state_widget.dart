import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

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
    final displayMessage = friendlyMessage(message);

    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 64,
              height: 64,
              decoration: BoxDecoration(
                color: AppColors.error.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.error_outline_rounded,
                color: AppColors.error,
                size: 32,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              displayMessage,
              style: AppTextStyles.body,
              textAlign: TextAlign.center,
            ),
            if (onRetry != null) ...[
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: onRetry,
                  child: const Text('Повторить'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
