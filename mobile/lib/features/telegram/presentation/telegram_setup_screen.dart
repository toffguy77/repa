import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../groups/presentation/groups_notifier.dart';
import 'connect_instruction_sheet.dart';
import 'telegram_notifier.dart';
import '../../../core/theme/app_tokens.dart';

class TelegramSetupScreen extends ConsumerStatefulWidget {
  final String groupId;

  const TelegramSetupScreen({super.key, required this.groupId});

  @override
  ConsumerState<TelegramSetupScreen> createState() =>
      _TelegramSetupScreenState();
}

class _TelegramSetupScreenState extends ConsumerState<TelegramSetupScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final groupState = ref.read(groupDetailProvider(widget.groupId));
      final username = groupState.detail?.group.telegramUsername;
      ref
          .read(telegramSetupProvider(widget.groupId).notifier)
          .init(telegramUsername: username);
    });
  }

  void _showConnectSheet() async {
    final notifier =
        ref.read(telegramSetupProvider(widget.groupId).notifier);
    await notifier.generateCode();
    final state = ref.read(telegramSetupProvider(widget.groupId));
    if (state.connectCode != null && mounted) {
      showModalBottomSheet(
        context: context,
        isScrollControlled: true,
        backgroundColor: Colors.transparent,
        builder: (_) => ConnectInstructionSheet(
          groupId: widget.groupId,
        ),
      );
    }
  }

  void _confirmDisconnect() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Отвязать Telegram?'),
        content: const Text(
          'Бот перестанет публиковать результаты в чат группы.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Отмена'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(ctx);
              ref
                  .read(telegramSetupProvider(widget.groupId).notifier)
                  .disconnect()
                  .then((_) {
                // Refresh group detail so telegramUsername is cleared
                ref
                    .read(groupDetailProvider(widget.groupId).notifier)
                    .load();
              });
            },
            style: TextButton.styleFrom(foregroundColor: context.t.color.danger),
            child: const Text('Отвязать'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(telegramSetupProvider(widget.groupId));

    ref.listen<TelegramSetupState>(
      telegramSetupProvider(widget.groupId),
      (prev, next) {
        if (next.error != null && prev?.error != next.error) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(next.error!)),
          );
          ref
              .read(telegramSetupProvider(widget.groupId).notifier)
              .clearError();
        }
      },
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Telegram')),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: state.connected ? _buildConnected(state) : _buildNotConnected(state),
      ),
    );
  }

  Widget _buildNotConnected(TelegramSetupState state) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Container(
          width: 100,
          height: 100,
          decoration: BoxDecoration(
            color: context.t.color.accentFill.withValues(alpha: 0.18),
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Icon(Icons.telegram, size: 56, color: context.t.color.accent),
          ),
        ),
        const SizedBox(height: 24),
        Text(
          'Подключите Telegram',
          style: context.ts.heading2,
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 12),
        Text(
          'Бот будет автоматически публиковать результаты голосования и анонсы новых сезонов в ваш Telegram-чат.',
          style: context.ts.bodySecondary,
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 32),
        SizedBox(
          width: double.infinity,
          height: 52,
          child: ElevatedButton.icon(
            onPressed: state.loading ? null : () {
              HapticFeedback.mediumImpact();
              _showConnectSheet();
            },
            icon: state.loading
                ? SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: context.t.color.onAccentFill,
                    ),
                  )
                : const Icon(Icons.telegram),
            label: Text(
              state.loading ? 'Загрузка...' : 'Подключить Telegram',
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: context.t.color.accent,
              foregroundColor: context.t.color.onAccentFill,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(AppTokens.radius.md),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildConnected(TelegramSetupState state) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Container(
          width: 80,
          height: 80,
          decoration: BoxDecoration(
            color: context.t.color.success.withValues(alpha: 0.14),
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Icon(Icons.check_circle, size: 48, color: context.t.color.success),
          ),
        ),
        const SizedBox(height: 20),
        Text(
          'Telegram подключён',
          style: context.ts.heading2,
        ),
        if (state.chatUsername != null) ...[
          const SizedBox(height: 8),
          Text(
            '@${state.chatUsername}',
            style: context.ts.bodySecondary,
          ),
        ],
        const SizedBox(height: 32),
        SizedBox(
          width: double.infinity,
          height: 48,
          child: OutlinedButton(
            onPressed: state.disconnecting ? null : () {
              HapticFeedback.mediumImpact();
              _confirmDisconnect();
            },
            style: OutlinedButton.styleFrom(
              foregroundColor: context.t.color.danger,
              side: BorderSide(color: context.t.color.danger),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(AppTokens.radius.md),
              ),
            ),
            child: state.disconnecting
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Отвязать'),
          ),
        ),
      ],
    );
  }
}
