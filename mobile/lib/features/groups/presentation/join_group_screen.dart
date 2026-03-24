import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../../core/analytics/analytics_service.dart';
import 'groups_notifier.dart';

class JoinGroupScreen extends ConsumerStatefulWidget {
  final String? initialCode;

  const JoinGroupScreen({super.key, this.initialCode});

  @override
  ConsumerState<JoinGroupScreen> createState() => _JoinGroupScreenState();
}

class _JoinGroupScreenState extends ConsumerState<JoinGroupScreen> {
  final _controller = TextEditingController();
  Timer? _debounce;

  @override
  void initState() {
    super.initState();
    if (widget.initialCode != null && widget.initialCode!.isNotEmpty) {
      _controller.text = widget.initialCode!;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ref.read(joinGroupProvider.notifier).loadPreview(widget.initialCode!);
      });
    }
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _controller.dispose();
    super.dispose();
  }

  void _onChanged(String value) {
    _debounce?.cancel();
    if (value.trim().isEmpty) {
      ref.read(joinGroupProvider.notifier).reset();
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 500), () {
      ref.read(joinGroupProvider.notifier).loadPreview(value);
    });
  }

  Future<void> _join() async {
    HapticFeedback.mediumImpact();
    final group = await ref
        .read(joinGroupProvider.notifier)
        .join(_controller.text);
    if (group != null && mounted) {
      ref.read(analyticsProvider).logGroupJoined();
      ref.read(groupsListProvider.notifier).refresh();
      context.go('/groups/${group.id}');
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(joinGroupProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Вступить в группу')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Инвайт-код или ссылка', style: AppTextStyles.body),
            const SizedBox(height: 8),
            TextField(
              controller: _controller,
              decoration: const InputDecoration(
                hintText: 'Вставь ссылку или код',
                prefixIcon: Icon(Icons.link),
              ),
              onChanged: _onChanged,
            ),
            const SizedBox(height: 16),
            if (state.previewing)
              const Center(child: CircularProgressIndicator()),
            if (state.preview != null)
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: AppColors.surface,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      state.preview!.name,
                      style: AppTextStyles.headline2.copyWith(fontSize: 18),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${state.preview!.memberCount} участников',
                      style: AppTextStyles.caption,
                    ),
                    Text(
                      'Админ: ${state.preview!.adminUsername}',
                      style: AppTextStyles.caption,
                    ),
                  ],
                ),
              ),
            if (state.error != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  state.error!,
                  style: AppTextStyles.caption.copyWith(color: AppColors.error),
                ),
              ),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              height: 50,
              child: ElevatedButton(
                onPressed: state.preview != null && !state.loading
                    ? _join
                    : null,
                child: state.loading
                    ? const SizedBox(
                        width: 24,
                        height: 24,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Text('Вступить'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
