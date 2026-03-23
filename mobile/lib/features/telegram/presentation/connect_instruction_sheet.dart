import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../groups/presentation/groups_notifier.dart';
import 'telegram_notifier.dart';

class ConnectInstructionSheet extends ConsumerStatefulWidget {
  final String groupId;

  const ConnectInstructionSheet({super.key, required this.groupId});

  @override
  ConsumerState<ConnectInstructionSheet> createState() =>
      _ConnectInstructionSheetState();
}

class _ConnectInstructionSheetState
    extends ConsumerState<ConnectInstructionSheet> {
  Timer? _countdownTimer;
  Duration _remaining = Duration.zero;
  bool _copied = false;

  @override
  void initState() {
    super.initState();
    _startCountdown();
  }

  void _startCountdown() {
    final state = ref.read(telegramSetupProvider(widget.groupId));
    final code = state.connectCode;
    if (code == null) return;

    final expiry = DateTime.parse(code.expiresAt);
    _remaining = expiry.difference(DateTime.now());
    if (_remaining.isNegative) _remaining = Duration.zero;

    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      setState(() {
        _remaining -= const Duration(seconds: 1);
        if (_remaining.isNegative) {
          _remaining = Duration.zero;
          _countdownTimer?.cancel();
        }
      });
    });
  }

  String _formatDuration(Duration d) {
    final hours = d.inHours;
    final minutes = d.inMinutes.remainder(60);
    final seconds = d.inSeconds.remainder(60);
    if (hours > 0) {
      return '$hoursч ${minutes.toString().padLeft(2, '0')}м';
    }
    return '$minutesм ${seconds.toString().padLeft(2, '0')}с';
  }

  void _copyCode(String code) {
    Clipboard.setData(ClipboardData(text: code));
    HapticFeedback.lightImpact();
    setState(() => _copied = true);
    Future.delayed(const Duration(seconds: 2), () {
      if (mounted) setState(() => _copied = false);
    });
  }

  void _openTelegram() {
    launchUrl(
      Uri.parse('tg://'),
      mode: LaunchMode.externalApplication,
    );
  }

  Future<void> _verify() async {
    HapticFeedback.mediumImpact();
    final connected = await ref
        .read(telegramSetupProvider(widget.groupId).notifier)
        .verifyConnection();
    if (connected && mounted) {
      // Refresh group detail
      ref.read(groupDetailProvider(widget.groupId).notifier).load();
      Navigator.pop(context);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Telegram подключён!')),
      );
    }
  }

  @override
  void dispose() {
    _countdownTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(telegramSetupProvider(widget.groupId));
    final code = state.connectCode;
    if (code == null) return const SizedBox.shrink();

    final expired = _remaining == Duration.zero;

    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
          const SizedBox(height: 20),
          Text('Как подключить', style: AppTextStyles.headline2),
          const SizedBox(height: 16),
          _buildStep('1', 'Добавьте @repaapp_bot в ваш Telegram-чат'),
          const SizedBox(height: 12),
          _buildStep('2', 'Сделайте бота администратором'),
          const SizedBox(height: 12),
          _buildStep('3', 'Напишите в чат:'),
          const SizedBox(height: 12),

          // Code display
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.primaryLight),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    '/connect ${code.connectCode}',
                    style: AppTextStyles.body.copyWith(
                      fontWeight: FontWeight.w600,
                      fontFamily: 'monospace',
                    ),
                  ),
                ),
                GestureDetector(
                  onTap: () => _copyCode('/connect ${code.connectCode}'),
                  child: AnimatedSwitcher(
                    duration: const Duration(milliseconds: 200),
                    child: _copied
                        ? const Icon(Icons.check,
                            key: ValueKey('check'),
                            color: AppColors.success,
                            size: 22)
                        : const Icon(Icons.copy,
                            key: ValueKey('copy'),
                            color: AppColors.primary,
                            size: 22),
                  ),
                ),
              ],
            ),
          ),

          const SizedBox(height: 12),

          // Countdown
          Center(
            child: Text(
              expired
                  ? 'Код истёк'
                  : 'Код действителен: ${_formatDuration(_remaining)}',
              style: AppTextStyles.caption.copyWith(
                color: expired ? AppColors.error : AppColors.textSecondary,
              ),
            ),
          ),

          const SizedBox(height: 20),

          // Action buttons
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: _openTelegram,
                  icon: const Icon(Icons.open_in_new, size: 18),
                  label: const Text('Telegram'),
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
              const SizedBox(width: 12),
              Expanded(
                child: ElevatedButton(
                  onPressed: (state.loading || expired) ? null : _verify,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primary,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  child: state.loading
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Text('Проверить'),
                ),
              ),
            ],
          ),

          if (state.error != null) ...[
            const SizedBox(height: 12),
            Text(
              state.error!,
              style: AppTextStyles.caption.copyWith(color: AppColors.error),
              textAlign: TextAlign.center,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildStep(String number, String text) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 24,
          height: 24,
          decoration: BoxDecoration(
            color: AppColors.primaryLight,
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              number,
              style: AppTextStyles.caption.copyWith(
                color: AppColors.primary,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.only(top: 2),
            child: Text(text, style: AppTextStyles.body),
          ),
        ),
      ],
    );
  }
}
