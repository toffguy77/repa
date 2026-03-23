import 'dart:async';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

const _weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

class RevealCountdownWidget extends StatefulWidget {
  final String revealAt;

  const RevealCountdownWidget({super.key, required this.revealAt});

  @override
  State<RevealCountdownWidget> createState() => _RevealCountdownWidgetState();
}

class _RevealCountdownWidgetState extends State<RevealCountdownWidget> {
  Timer? _timer;
  Duration _remaining = Duration.zero;
  DateTime? _revealTime;

  @override
  void initState() {
    super.initState();
    _revealTime = DateTime.tryParse(widget.revealAt)?.toLocal();
    _updateRemaining();
    _timer = Timer.periodic(const Duration(seconds: 30), (_) => _updateRemaining());
  }

  @override
  void didUpdateWidget(RevealCountdownWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.revealAt != widget.revealAt) {
      _revealTime = DateTime.tryParse(widget.revealAt)?.toLocal();
      _updateRemaining();
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _updateRemaining() {
    if (_revealTime == null) return;
    setState(() {
      _remaining = _revealTime!.difference(DateTime.now());
      if (_remaining.isNegative) _remaining = Duration.zero;
    });
  }

  String _formatRemaining() {
    if (_remaining == Duration.zero) return 'Скоро!';

    final days = _remaining.inDays;
    final hours = _remaining.inHours % 24;
    final minutes = _remaining.inMinutes % 60;

    if (days > 0) return '$daysд $hoursч';
    if (hours > 0) return '$hoursч $minutesмин';
    return '$minutesмин';
  }

  String _formatRevealDay() {
    if (_revealTime == null) return '';
    final day = _weekdays[_revealTime!.weekday - 1];
    final time = DateFormat('HH:mm').format(_revealTime!);
    return '$day, $time';
  }

  @override
  Widget build(BuildContext context) {
    final isUrgent = _remaining.inHours < 1 && _remaining > Duration.zero;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: isUrgent
            ? AppColors.error.withValues(alpha: 0.1)
            : AppColors.primaryLight,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        '${_formatRevealDay()} \u00b7 ${_formatRemaining()}',
        style: AppTextStyles.caption.copyWith(
          fontSize: 12,
          color: isUrgent ? AppColors.error : AppColors.primary,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}
