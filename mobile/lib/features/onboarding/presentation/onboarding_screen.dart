import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../core/theme/app_tokens.dart';

/// Onboarding is intentionally always dark — it is the product's first impression and
/// matches the shared reputation card. It therefore names the dark palette directly rather
/// than following the active theme. See docs/features/design-system.md.
const _palette = AppColorTokens.dark;
final _bgColor = _palette.canvas;

class OnboardingScreen extends ConsumerStatefulWidget {
  const OnboardingScreen({super.key});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  final _controller = PageController();
  int _currentPage = 0;

  static const _slides = [
    _SlideData(
      emoji: '\u{1F346}',
      title: 'Создай группу\nдля своих',
      subtitle: 'Добавь класс, компанию\nили просто друзей\nпо ссылке-инвайту',
    ),
    _SlideData(
      emoji: '\u{1F5F3}',
      title: 'Голосуй анонимно',
      subtitle:
          'Каждую неделю — смешные\nвопросы про участников.\nНикто не узнает твой ответ.',
    ),
    _SlideData(
      emoji: '\u{1F3AD}',
      title: 'Узнай в пятницу',
      subtitle:
          'В 20:00 — твоя карточка\nрепутации. Поделись\nс чатом или оставь себе.',
    ),
  ];

  Future<void> _finish() async {
    HapticFeedback.mediumImpact();
    ref.read(authProvider.notifier).onboardingCompleted();
    if (mounted) context.go('/home');
  }

  void _skip() => _finish();

  void _next() {
    if (_currentPage < _slides.length - 1) {
      HapticFeedback.lightImpact();
      _controller.nextPage(
        duration: context.motion(MotionClass.emphasis),
        curve: Curves.easeOutCubic,
      );
    } else {
      _finish();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _bgColor,
      body: SafeArea(
        child: Column(
          children: [
            // Skip button
            Align(
              alignment: Alignment.topRight,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: TextButton(
                  onPressed: _skip,
                  child: Text(
                    'Пропустить',
                    style: TextStyle(
                      color: _palette.textSecondary,
                      fontSize: 16,
                    ),
                  ),
                ),
              ),
            ),

            // Slides
            Expanded(
              child: PageView.builder(
                controller: _controller,
                itemCount: _slides.length,
                onPageChanged: (i) => setState(() => _currentPage = i),
                itemBuilder: (context, i) => _SlideWidget(data: _slides[i]),
              ),
            ),

            // Dots
            Padding(
              padding: const EdgeInsets.only(bottom: 24),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: List.generate(
                  _slides.length,
                  (i) => AnimatedContainer(
                    duration: context.motion(MotionClass.surface),
                    margin: const EdgeInsets.symmetric(horizontal: 4),
                    width: _currentPage == i ? 24 : 8,
                    height: 8,
                    decoration: BoxDecoration(
                      color: _currentPage == i
                          ? _palette.accent
                          : _palette.border,
                      borderRadius: BorderRadius.circular(AppTokens.radius.xs),
                    ),
                  ),
                ),
              ),
            ),

            // Button
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 0, 24, 32),
              child: SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton(
                  onPressed: _next,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: context.t.color.accent,
                    foregroundColor: _palette.onAccentFill,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(AppTokens.radius.lg),
                    ),
                    textStyle: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  child: Text(
                    _currentPage == _slides.length - 1
                        ? 'Начать'
                        : 'Дальше',
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _SlideData {
  final String emoji;
  final String title;
  final String subtitle;

  const _SlideData({
    required this.emoji,
    required this.title,
    required this.subtitle,
  });
}

class _SlideWidget extends StatelessWidget {
  final _SlideData data;

  const _SlideWidget({required this.data});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            data.emoji,
            style: const TextStyle(fontSize: 96),
          )
              .animate()
              .fadeIn(duration: context.motion(MotionClass.emphasis))
              .scale(begin: const Offset(0.5, 0.5), end: const Offset(1, 1)),
          const SizedBox(height: 32),
          Text(
            data.title,
            textAlign: TextAlign.center,
            style: AppTokens.text.heading1.copyWith(color: _palette.textPrimary),
          ).animate().fadeIn(delay: AppTokens.motion.stagger(2), duration: context.motion(MotionClass.emphasis)).slideY(
                begin: 0.2,
                end: 0,
                curve: Curves.easeOut,
              ),
          const SizedBox(height: 16),
          Text(
            data.subtitle,
            textAlign: TextAlign.center,
            style: TextStyle(
              color: _palette.textSecondary,
              fontSize: 16,
              height: 1.5,
            ),
          ).animate().fadeIn(delay: AppTokens.motion.stagger(3), duration: context.motion(MotionClass.emphasis)).slideY(
                begin: 0.2,
                end: 0,
                curve: Curves.easeOut,
              ),
        ],
      ),
    );
  }
}
