import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../core/theme/app_tokens.dart';


class PaymentPendingSheet extends StatelessWidget {
  final VoidCallback onConfirm;

  const PaymentPendingSheet({super.key, required this.onConfirm});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: context.t.elevation.level2.surface,
        borderRadius: AppTokens.radius.sheet,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: context.t.color.surface,
              borderRadius: BorderRadius.circular(AppTokens.radius.xs),
            ),
          ),
          const SizedBox(height: 24),
          SizedBox(
            width: 48,
            height: 48,
            child: CircularProgressIndicator(
              strokeWidth: 3,
              color: context.t.color.accent,
            ),
          ),
          const SizedBox(height: 20),
          Text('Ожидание оплаты...', style: context.ts.heading2),
          const SizedBox(height: 8),
          Text(
            'Завершите оплату в браузере и вернитесь в приложение',
            style: context.ts.caption,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          Text(
            'Внешний способ оплаты',
            style: context.ts.caption.copyWith(
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 24),
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton(
              onPressed: () {
                HapticFeedback.mediumImpact();
                onConfirm();
              },
              child: const Text('Я оплатил'),
            ),
          ),
          const SizedBox(height: 12),
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(
              'Отмена',
              style: TextStyle(color: context.t.color.textSecondary),
            ),
          ),
          const SizedBox(height: 8),
        ],
      ),
    );
  }
}
