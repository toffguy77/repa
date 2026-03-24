import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../core/theme/app_colors.dart';

const _bgColor = Color(0xFF1A0533);

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
        duration: const Duration(milliseconds: 400),
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
                      color: Colors.white.withValues(alpha: 0.6),
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
                    duration: const Duration(milliseconds: 300),
                    margin: const EdgeInsets.symmetric(horizontal: 4),
                    width: _currentPage == i ? 24 : 8,
                    height: 8,
                    decoration: BoxDecoration(
                      color: _currentPage == i
                          ? Colors.white
                          : Colors.white.withValues(alpha: 0.3),
                      borderRadius: BorderRadius.circular(4),
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
                    backgroundColor: AppColors.primary,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
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
              .fadeIn(duration: 400.ms)
              .scale(begin: const Offset(0.5, 0.5), end: const Offset(1, 1)),
          const SizedBox(height: 32),
          Text(
            data.title,
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.bold,
              height: 1.3,
            ),
          ).animate().fadeIn(delay: 100.ms, duration: 400.ms).slideY(
                begin: 0.2,
                end: 0,
                curve: Curves.easeOut,
              ),
          const SizedBox(height: 16),
          Text(
            data.subtitle,
            textAlign: TextAlign.center,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.7),
              fontSize: 16,
              height: 1.5,
            ),
          ).animate().fadeIn(delay: 200.ms, duration: 400.ms).slideY(
                begin: 0.2,
                end: 0,
                curve: Curves.easeOut,
              ),
        ],
      ),
    );
  }
}
