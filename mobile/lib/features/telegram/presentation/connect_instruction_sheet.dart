import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../groups/presentation/groups_notifier.dart';
import 'telegram_notifier.dart';
import '../../../core/theme/app_tokens.dart';

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
      decoration: BoxDecoration(
        color: context.t.elevation.level2.surface,
        borderRadius: AppTokens.radius.sheet,
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
                borderRadius: BorderRadius.circular(AppTokens.radius.xs),
              ),
            ),
          ),
          const SizedBox(height: 20),
          Text('Как подключить', style: context.ts.heading2),
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
              color: context.t.color.surface,
              borderRadius: BorderRadius.circular(AppTokens.radius.md),
              border: Border.all(color: context.t.color.accentFill.withValues(alpha: 0.18)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    '/connect ${code.connectCode}',
                    style: context.ts.body.copyWith(
                      fontWeight: FontWeight.w600,
                      fontFamily: 'monospace',
                    ),
                  ),
                ),
                GestureDetector(
                  onTap: () => _copyCode('/connect ${code.connectCode}'),
                  child: AnimatedSwitcher(
                    duration: context.motion(MotionClass.surface),
                    child: _copied
                        ? Icon(Icons.check,
                            key: ValueKey('check'),
                            color: context.t.color.success,
                            size: 22)
                        : Icon(Icons.copy,
                            key: ValueKey('copy'),
                            color: context.t.color.accent,
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
              style: context.ts.caption.copyWith(
                color: expired ? context.t.color.danger : context.t.color.textSecondary,
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
                    foregroundColor: context.t.color.accent,
                    side: BorderSide(color: context.t.color.accent),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(AppTokens.radius.md),
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
                    backgroundColor: context.t.color.accent,
                    foregroundColor: context.t.color.onAccentFill,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(AppTokens.radius.md),
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  child: state.loading
                      ? SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: context.t.color.onAccentFill,
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
              style: context.ts.caption.copyWith(color: context.t.color.danger),
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
            color: context.t.color.accentFill.withValues(alpha: 0.18),
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              number,
              style: context.ts.caption.copyWith(
                color: context.t.color.accent,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.only(top: 2),
            child: Text(text, style: context.ts.body),
          ),
        ),
      ],
    );
  }
}
