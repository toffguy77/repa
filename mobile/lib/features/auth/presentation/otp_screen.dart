import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:pinput/pinput.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../../core/analytics/analytics_service.dart';
import 'auth_notifier.dart';

class OtpScreen extends ConsumerStatefulWidget {
  final String phone;

  const OtpScreen({super.key, required this.phone});

  @override
  ConsumerState<OtpScreen> createState() => _OtpScreenState();
}

class _OtpScreenState extends ConsumerState<OtpScreen> {
  final _pinController = TextEditingController();
  final _focusNode = FocusNode();
  Timer? _timer;
  int _secondsLeft = 300; // 5 minutes

  @override
  void initState() {
    super.initState();
    _startTimer();
  }

  void _startTimer() {
    _secondsLeft = 300;
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_secondsLeft <= 0) {
        timer.cancel();
      } else {
        setState(() => _secondsLeft--);
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    _pinController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  String get _formattedTime {
    final minutes = _secondsLeft ~/ 60;
    final seconds = _secondsLeft % 60;
    return '${minutes.toString().padLeft(2, '0')}:${seconds.toString().padLeft(2, '0')}';
  }

  Future<void> _verify(String code) async {
    final success =
        await ref.read(otpVerifyProvider.notifier).verify(widget.phone, code);
    if (success) {
      ref.read(analyticsProvider).logLogin('phone');
    }
  }

  Future<void> _resend() async {
    ref.read(otpVerifyProvider.notifier).reset();
    final success =
        await ref.read(otpSendProvider.notifier).send(widget.phone);
    if (success) {
      _pinController.clear();
      _startTimer();
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(otpVerifyProvider);

    final defaultPinTheme = PinTheme(
      width: 48,
      height: 56,
      textStyle: const TextStyle(
        fontSize: 22,
        fontWeight: FontWeight.w600,
        color: AppColors.textPrimary,
      ),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(12),
      ),
    );

    return Scaffold(
      appBar: AppBar(
        leading: const BackButton(),
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 24),
              Text('Введи код', style: AppTextStyles.headline1),
              const SizedBox(height: 8),
              Text(
                'Отправили SMS на ${widget.phone}',
                style: AppTextStyles.bodySecondary,
              ),
              const SizedBox(height: 32),
              Center(
                child: Pinput(
                  length: 6,
                  controller: _pinController,
                  focusNode: _focusNode,
                  autofocus: true,
                  defaultPinTheme: defaultPinTheme,
                  focusedPinTheme: defaultPinTheme.copyWith(
                    decoration: defaultPinTheme.decoration!.copyWith(
                      border:
                          Border.all(color: AppColors.primary, width: 2),
                    ),
                  ),
                  errorPinTheme: defaultPinTheme.copyWith(
                    decoration: defaultPinTheme.decoration!.copyWith(
                      border: Border.all(color: AppColors.error),
                    ),
                  ),
                  enabled: !state.loading,
                  onCompleted: _verify,
                ),
              ),
              if (state.error != null) ...[
                const SizedBox(height: 12),
                Center(
                  child: Text(
                    state.error!,
                    style: TextStyle(color: AppColors.error, fontSize: 14),
                  ),
                ),
              ],
              const SizedBox(height: 24),
              if (state.loading)
                const Center(child: CircularProgressIndicator())
              else
                Center(
                  child: _secondsLeft > 0
                      ? Text(
                          'Отправить повторно через $_formattedTime',
                          style: AppTextStyles.caption,
                        )
                      : TextButton(
                          onPressed: _resend,
                          child: const Text(
                            'Отправить код повторно',
                            style: TextStyle(color: AppColors.primary),
                          ),
                        ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
