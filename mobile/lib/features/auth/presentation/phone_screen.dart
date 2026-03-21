import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mask_text_input_formatter/mask_text_input_formatter.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import 'auth_notifier.dart';

class PhoneScreen extends ConsumerStatefulWidget {
  const PhoneScreen({super.key});

  @override
  ConsumerState<PhoneScreen> createState() => _PhoneScreenState();
}

class _PhoneScreenState extends ConsumerState<PhoneScreen> {
  final _controller = TextEditingController();
  final _formatter = MaskTextInputFormatter(
    mask: '+7 (###) ###-##-##',
    filter: {'#': RegExp(r'[0-9]')},
  );

  String get _rawPhone {
    final digits = _formatter.getUnmaskedText();
    return '+7$digits';
  }

  bool get _isValid => _formatter.getUnmaskedText().length == 10;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_isValid) return;
    final success = await ref.read(otpSendProvider.notifier).send(_rawPhone);
    if (success && mounted) {
      context.push('/auth/otp', extra: _rawPhone);
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(otpSendProvider);

    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 60),
              Text('Вход в Репу', style: AppTextStyles.headline1),
              const SizedBox(height: 8),
              Text(
                'Введи номер телефона, чтобы получить код',
                style: AppTextStyles.bodySecondary,
              ),
              const SizedBox(height: 32),
              TextField(
                controller: _controller,
                inputFormatters: [_formatter],
                keyboardType: TextInputType.phone,
                style: AppTextStyles.body,
                decoration: const InputDecoration(
                  hintText: '+7 (___) ___-__-__',
                ),
                onChanged: (_) => setState(() {}),
              ),
              if (state.error != null) ...[
                const SizedBox(height: 8),
                Text(state.error!, style: TextStyle(color: AppColors.error)),
              ],
              const SizedBox(height: 24),
              ElevatedButton(
                onPressed: _isValid && !state.loading ? _submit : null,
                child: state.loading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Text('Получить код'),
              ),
              const Spacer(),
              _SocialButtons(),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }
}

class _SocialButtons extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text('или', style: AppTextStyles.caption),
        const SizedBox(height: 16),
        OutlinedButton.icon(
          onPressed: null, // stub
          icon: const Icon(Icons.apple),
          label: const Text('Войти через Apple'),
        ),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          onPressed: null, // stub
          icon: const Icon(Icons.g_mobiledata),
          label: const Text('Войти через Google'),
        ),
      ],
    );
  }
}
