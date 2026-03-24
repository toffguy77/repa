import 'package:freezed_annotation/freezed_annotation.dart';

part 'question_candidate.freezed.dart';
part 'question_candidate.g.dart';

@freezed
class QuestionCandidate with _$QuestionCandidate {
  const factory QuestionCandidate({
    required String id,
    required String text,
    required String category,
  }) = _QuestionCandidate;

  factory QuestionCandidate.fromJson(Map<String, dynamic> json) =>
      _$QuestionCandidateFromJson(json);
}
