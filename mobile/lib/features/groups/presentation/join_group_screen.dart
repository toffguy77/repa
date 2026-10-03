import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/analytics/analytics_service.dart';
import 'groups_notifier.dart';
import '../../../core/theme/app_tokens.dart';
import 'widgets/invite_share_sheet.dart';
import '../../reveal/domain/share_link.dart';

class JoinGroupScreen extends ConsumerStatefulWidget {
  final String? initialCode;

  /// How the user arrived. The join screen itself is the typed-code path; a deep link passes
  /// the source its channel marker implies.
  final JoinSource source;

  /// The member who shared the link, so a referral reward reaches them. Validated server-side;
  /// an invalid claim is dropped rather than failing the join.
  final String? referrerId;

  const JoinGroupScreen({
    super.key,
    this.initialCode,
    this.source = JoinSource.code,
    this.referrerId,
  });

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
      // The code travels as a URL path segment, so a pasted link has to be reduced to the
      // bare code here; the server still normalises case and separators.
      ref.read(joinGroupProvider.notifier).loadPreview(extractInviteCode(value));
    });
  }

  Future<void> _join() async {
    HapticFeedback.mediumImpact();
    final group = await ref
        .read(joinGroupProvider.notifier)
        .join(
          extractInviteCode(_controller.text),
          source: widget.source,
          referrerId: widget.referrerId,
        );
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
            Text('Код группы или ссылка', style: context.ts.body),
            const SizedBox(height: 8),
            TextField(
              controller: _controller,
              textCapitalization: TextCapitalization.characters,
              autocorrect: false,
              // Codes are case-insensitive server-side; upper-casing as you type is purely
              // so what you see matches the code you were given.
              inputFormatters: [UpperCaseFormatter()],
              decoration: const InputDecoration(
                hintText: 'AB2 CD3 или ссылка',
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
                  color: context.t.color.surface,
                  borderRadius: BorderRadius.circular(AppTokens.radius.md),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      state.preview!.name,
                      style: context.ts.heading2.copyWith(fontSize: 18),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${state.preview!.memberCount} участников',
                      style: context.ts.caption,
                    ),
                    Text(
                      'Админ: ${state.preview!.adminUsername}',
                      style: context.ts.caption,
                    ),
                  ],
                ),
              ),
            if (state.error != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  state.error!,
                  style: context.ts.caption.copyWith(color: context.t.color.danger),
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
                    ? SizedBox(
                        width: 24,
                        height: 24,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: context.t.color.onAccentFill,
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

/// Upper-cases as the user types. Display only — the backend normalises authoritatively
/// (see docs/features/groups.md → Invite codes).
class UpperCaseFormatter extends TextInputFormatter {
  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    // Leave a pasted URL alone: upper-casing a host or path can break it.
    if (newValue.text.contains('/')) return newValue;
    return newValue.copyWith(text: newValue.text.toUpperCase());
  }
}
