import 'package:flutter/material.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';

class StatCard extends StatefulWidget {
  final String label;
  final String value;
  final IconData icon;
  final bool animateNumber;

  const StatCard({
    super.key,
    required this.label,
    required this.value,
    required this.icon,
    this.animateNumber = true,
  });

  @override
  State<StatCard> createState() => _StatCardState();
}

class _StatCardState extends State<StatCard>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 800),
      vsync: this,
    );
    _animation = CurvedAnimation(parent: _controller, curve: Curves.easeOut);

    if (widget.animateNumber) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _controller.forward();
      });
    } else {
      _controller.value = 1.0;
    }
  }

  @override
  void didUpdateWidget(StatCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.value != widget.value && widget.animateNumber) {
      _controller.forward(from: 0);
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final numValue = double.tryParse(
        widget.value.replaceAll('%', '').replaceAll(',', '.'));

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(widget.icon, size: 20, color: AppColors.primary),
          const SizedBox(height: 8),
          if (numValue != null && widget.animateNumber)
            AnimatedBuilder(
              animation: _animation,
              builder: (context, _) {
                final current = numValue * _animation.value;
                final display = widget.value.contains('%')
                    ? '${current.toStringAsFixed(1)}%'
                    : current.toInt().toString();
                return Text(
                  display,
                  style: AppTextStyles.headline2.copyWith(fontSize: 20),
                );
              },
            )
          else
            Text(
              widget.value,
              style: AppTextStyles.headline2.copyWith(fontSize: 20),
            ),
          const SizedBox(height: 4),
          Text(
            widget.label,
            style: AppTextStyles.caption.copyWith(fontSize: 12),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }
}

