import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../groups/presentation/groups_notifier.dart';
import 'connect_instruction_sheet.dart';
import 'telegram_notifier.dart';

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
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
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
            color: AppColors.primaryLight,
            shape: BoxShape.circle,
          ),
          child: const Center(
            child: Icon(Icons.telegram, size: 56, color: AppColors.primary),
          ),
        ),
        const SizedBox(height: 24),
        Text(
          'Подключите Telegram',
          style: AppTextStyles.headline2,
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 12),
        Text(
          'Бот будет автоматически публиковать результаты голосования и анонсы новых сезонов в ваш Telegram-чат.',
          style: AppTextStyles.bodySecondary,
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
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : const Icon(Icons.telegram),
            label: Text(
              state.loading ? 'Загрузка...' : 'Подключить Telegram',
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
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
          decoration: const BoxDecoration(
            color: Color(0xFFD1FAE5),
            shape: BoxShape.circle,
          ),
          child: const Center(
            child: Icon(Icons.check_circle, size: 48, color: AppColors.success),
          ),
        ),
        const SizedBox(height: 20),
        Text(
          'Telegram подключён',
          style: AppTextStyles.headline2,
        ),
        if (state.chatUsername != null) ...[
          const SizedBox(height: 8),
          Text(
            '@${state.chatUsername}',
            style: AppTextStyles.bodySecondary,
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
              foregroundColor: AppColors.error,
              side: const BorderSide(color: AppColors.error),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
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
