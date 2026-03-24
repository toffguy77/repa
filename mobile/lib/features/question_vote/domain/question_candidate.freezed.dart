// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'question_candidate.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

QuestionCandidate _$QuestionCandidateFromJson(Map<String, dynamic> json) {
  return _QuestionCandidate.fromJson(json);
}

/// @nodoc
mixin _$QuestionCandidate {
  String get id => throw _privateConstructorUsedError;
  String get text => throw _privateConstructorUsedError;
  String get category => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $QuestionCandidateCopyWith<QuestionCandidate> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $QuestionCandidateCopyWith<$Res> {
  factory $QuestionCandidateCopyWith(
          QuestionCandidate value, $Res Function(QuestionCandidate) then) =
      _$QuestionCandidateCopyWithImpl<$Res, QuestionCandidate>;
  @useResult
  $Res call({String id, String text, String category});
}

/// @nodoc
class _$QuestionCandidateCopyWithImpl<$Res, $Val extends QuestionCandidate>
    implements $QuestionCandidateCopyWith<$Res> {
  _$QuestionCandidateCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? text = null,
    Object? category = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      text: null == text
          ? _value.text
          : text // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$QuestionCandidateImplCopyWith<$Res>
    implements $QuestionCandidateCopyWith<$Res> {
  factory _$$QuestionCandidateImplCopyWith(_$QuestionCandidateImpl value,
          $Res Function(_$QuestionCandidateImpl) then) =
      __$$QuestionCandidateImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({String id, String text, String category});
}

/// @nodoc
class __$$QuestionCandidateImplCopyWithImpl<$Res>
    extends _$QuestionCandidateCopyWithImpl<$Res, _$QuestionCandidateImpl>
    implements _$$QuestionCandidateImplCopyWith<$Res> {
  __$$QuestionCandidateImplCopyWithImpl(_$QuestionCandidateImpl _value,
      $Res Function(_$QuestionCandidateImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? text = null,
    Object? category = null,
  }) {
    return _then(_$QuestionCandidateImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      text: null == text
          ? _value.text
          : text // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$QuestionCandidateImpl implements _QuestionCandidate {
  const _$QuestionCandidateImpl(
      {required this.id, required this.text, required this.category});

  factory _$QuestionCandidateImpl.fromJson(Map<String, dynamic> json) =>
      _$$QuestionCandidateImplFromJson(json);

  @override
  final String id;
  @override
  final String text;
  @override
  final String category;

  @override
  String toString() {
    return 'QuestionCandidate(id: $id, text: $text, category: $category)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$QuestionCandidateImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.text, text) || other.text == text) &&
            (identical(other.category, category) ||
                other.category == category));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, id, text, category);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$QuestionCandidateImplCopyWith<_$QuestionCandidateImpl> get copyWith =>
      __$$QuestionCandidateImplCopyWithImpl<_$QuestionCandidateImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$QuestionCandidateImplToJson(
      this,
    );
  }
}

abstract class _QuestionCandidate implements QuestionCandidate {
  const factory _QuestionCandidate(
      {required final String id,
      required final String text,
      required final String category}) = _$QuestionCandidateImpl;

  factory _QuestionCandidate.fromJson(Map<String, dynamic> json) =
      _$QuestionCandidateImpl.fromJson;

  @override
  String get id;
  @override
  String get text;
  @override
  String get category;
  @override
  @JsonKey(ignore: true)
  _$$QuestionCandidateImplCopyWith<_$QuestionCandidateImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
